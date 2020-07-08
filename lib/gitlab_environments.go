package gc

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/plouc/go-gitlab-client/gitlab"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
)

// GitlabEnvironments removes gitlab environments without a ingress definition in kubernetes
func GitlabEnvironments(ctx context.Context, n core_v1.NodeInterface, i v1beta1.IngressInterface, g *gitlab.Gitlab) error {

	nodeIPs, err := fetchNodeIPs(ctx, n)
	if err != nil {
		return fmt.Errorf("failed to fetch kubernetes node IPs: %s", err)
	}
	hostNames, err := fetchIngressHostnames(ctx, i)
	if err != nil {
		return fmt.Errorf("failed to fetch ingress host names: %s", err)
	}
	runnerIDs, err := fetchRunnerIds(nodeIPs, g)
	if err != nil {
		return fmt.Errorf("failed to fetch gitlab runner ids: %s", err)
	}
	for _, id := range runnerIDs {
		runner, _, err := g.Runner(id)
		if err != nil {
			return fmt.Errorf("failed to fetch runner details: %s", err)
		}
		err = removeGitlabEnvironments(runner.Projects, hostNames, g)
		if err != nil {
			return fmt.Errorf("failed to remove gitlab project environments: %s", err)
		}
	}
	return nil
}

func fetchRunnerIds(nodeIps map[string]bool, g *gitlab.Gitlab) ([]int, error) {
	runnersCollection, _, err := g.Runners(&gitlab.RunnersOptions{All: true, PaginationOptions: gitlab.PaginationOptions{PerPage: 100}})
	if err != nil {
		return nil, err
	}
	IDs := []int{}
	for _, runner := range runnersCollection.Items {
		if nodeIps[runner.IpAddress] {
			IDs = append(IDs, runner.Id)
		}
	}
	return IDs, nil
}

func fetchNodeIPs(ctx context.Context, n core_v1.NodeInterface) (map[string]bool, error) {
	nodes, err := n.List(ctx, meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	IPPattern := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	ips := map[string]bool{}
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if IPPattern.MatchString(address.Address) {
				ips[address.Address] = true
			}
		}
	}
	return ips, nil
}

func fetchIngressHostnames(ctx context.Context, i v1beta1.IngressInterface) (map[string]bool, error) {
	ingresses, err := i.List(ctx, meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	hosts := map[string]bool{}
	for _, ingress := range ingresses.Items {
		for _, rule := range ingress.Spec.Rules {
			if rule.Host != "" {
				hosts[rule.Host] = true
			}
		}
	}
	return hosts, nil
}

func removeGitlabEnvironments(projects []*gitlab.Project, hostNames map[string]bool, g *gitlab.Gitlab) error {
	for _, p := range projects {
		environments := []gitlab.Environment{}
		for environmentWithError := range generateEnvironments(g, strconv.Itoa(p.Id)) {
			if environmentWithError.err != nil {
				return environmentWithError.err
			}
			environments = append(environments, environmentWithError.environment)
		}
		for _, environment := range environments {
			err := removeGitlabEnvironment(environment, hostNames, g)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func removeGitlabEnvironment(e gitlab.Environment, hostNames map[string]bool, g *gitlab.Gitlab) error {
	if urlExists(e.ExternalUrl, hostNames) {
		return nil
	}
	fmt.Printf("deleting gitlab environment prj_id: %d %v - env_id %d %v\n", e.Project.Id, e.Project.NameWithNamespace, e.Id, e.ExternalUrl)
	_, err := g.RemoveProjectEnvironment(strconv.Itoa(e.Project.Id), e.Id)
	return err
}

func urlExists(externalURL string, ingresses map[string]bool) bool {
	url := ""
	if strings.HasPrefix(externalURL, "http://") {
		url = externalURL[7:]
	}
	if strings.HasPrefix(externalURL, "https://") {
		url = externalURL[8:]
	}
	return ingresses[url]
}

type environmentWithError struct {
	environment gitlab.Environment
	err         error
}

func generateEnvironments(g *gitlab.Gitlab, projectID string) chan environmentWithError {
	ch := make(chan environmentWithError)
	go func() {
		defer close(ch)
		page := 1
		for ok := true; ok; {
			environmentsPage, meta, err := g.ProjectEnvironments(projectID, &gitlab.PaginationOptions{Page: page, PerPage: 100})
			if err != nil {
				ch <- environmentWithError{
					err: err,
				}
				return
			}
			for _, environment := range environmentsPage.Items {
				ch <- environmentWithError{
					environment: *environment,
				}
			}
			ok = page != meta.TotalPages
			page++
		}
	}()
	return ch
}
