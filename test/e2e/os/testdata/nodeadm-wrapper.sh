#!/usr/bin/env bash

# Build nodeadm command with optional manifest override
NODEADM_ARGS=("$@")

{{- if .ManifestURL }}
# Add --manifest-override flag only for commands that support it
COMMAND="${1:-}"
if [[ "$COMMAND" == "init" || "$COMMAND" == "upgrade" || "$COMMAND" == "install" ]]; then
  NODEADM_ARGS=("--manifest-override" "{{ .ManifestURL }}" "${NODEADM_ARGS[@]}")
fi
{{- end }}

AWS_ENDPOINT_URL_EKS={{ .EKSEndpoint }} /usr/local/bin/nodeadm "${NODEADM_ARGS[@]}"
