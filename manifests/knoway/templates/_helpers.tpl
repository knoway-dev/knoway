{{/*
Return the proper image name
*/}}

{{- define "knoway.gateway.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global "defaultTag" .Chart.Version) }}
{{- end -}}
