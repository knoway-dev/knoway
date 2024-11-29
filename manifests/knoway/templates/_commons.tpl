{{/*
Common Template
*/}}

{{/*
Merge imagePullSecrets: common.images.pullSecrets
*/}}
{{- define "common.images.pullSecrets" -}}
    {{- $pullSecrets := list }}

    {{- if .Values.global -}}
        {{- range .Values.global.imagePullSecrets -}}
            {{- $pullSecrets = append $pullSecrets . -}}
        {{- end -}}
    {{- end -}}
    {{- range .Values.imagePullSecrets -}}
        {{- $pullSecrets = append $pullSecrets . -}}
    {{- end -}}

    {{- if (not (empty $pullSecrets)) }}
imagePullSecrets:
        {{- range $pullSecrets }}
  - name: {{ . }}
        {{- end }}
    {{- end }}
{{- end -}}

{{/*
Merge Resource: common.images.resources
*/}}
{{- define "common.images.resources" -}}

    {{- if .Values.resources }}
{{ toYaml .Values.resources }}
    {{- else if .Values.global }}
            {{- if .Values.global.resources }}
{{ toYaml .Values.global.resources }}
            {{- end }}
    {{- end }}

{{- end -}}

{{/*
Return the proper image name
Usage:     {{ include "common.images.image" ( dict "imageRoot" .imageRootPath "global" .globalPath "defaultTag" .tagPath) }}
*/}}
{{- define "common.images.image" -}}
{{- $registryName := .imageRoot.registry -}}
{{- $repositoryName := .imageRoot.repository -}}
{{- $tag := .defaultTag  -}}
{{- if .global }}
    {{- if .global.imageRegistry }}
     {{- $registryName = .global.imageRegistry -}}
    {{- end -}}
{{- end -}}
{{- if .imageRoot.registry }}
    {{- $registryName = .imageRoot.registry  -}}
{{- end -}}
{{- if .imageRoot.tag }}
    {{- $tag = .imageRoot.tag  -}}
{{- end -}}
{{- if $registryName }}
{{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- else -}}
{{- printf "%s:%s" $repositoryName $tag -}}
{{- end -}}
{{- end -}}

{{- define "replicas" -}}
    {{- if .Values.replicas }}{{.Values.replicas}}{{else}}{{ if .Values.global.high_available }}2{{else}}1{{end}}{{end -}}
{{- end -}}

{{- define "hpa.min_replicas" -}}
    {{- if .Values.global.high_available }}2{{else}}1{{end}}
{{- end -}}
