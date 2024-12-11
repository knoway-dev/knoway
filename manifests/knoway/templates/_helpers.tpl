{{/*
Return the proper image name
*/}}

{{- define "knoway.gateway.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.gateway.image "global" .Values.global "defaultTag" .Chart.Version) }}
{{- end -}}
