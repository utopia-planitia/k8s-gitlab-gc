# from tiltapi import *

disable_snapshots()
analytics_settings(enable=False)
allow_k8s_contexts(os.getenv("TILT_ALLOW_CONTEXT"))
if os.environ.get('TILT_REGISTRY_PUSH'):
    default_registry(os.environ.get('TILT_REGISTRY_PUSH'), os.environ.get('TILT_REGISTRY_PULL'))

include('./deployments/Tiltfile')

