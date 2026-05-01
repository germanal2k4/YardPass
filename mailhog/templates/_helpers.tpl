{{- define "mailhog.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "mailhog.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- include "mailhog.name" . -}}
{{- end -}}
{{- end -}}

{{- define "mailhog.labels" -}}
app.kubernetes.io/name: {{ include "mailhog.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
