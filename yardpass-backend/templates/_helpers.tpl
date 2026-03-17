{{- define "yardpass-backend.release-name" -}}
{{ .Release.Name }}
{{- end }}

{{- define "yardpass-backend.resource-name" -}}
{{- .Values.appLabel -}}
{{- end }}

{{- define "yardpass-backend.selector" -}}
app: {{ .Values.appLabel }}
{{- end }}

{{- define "yardpass-backend.labels" -}}
app: {{ .Values.appLabel }}
{{- end }}

{{- define "yardpass-backend.liveness" -}}
{{- if .Values.livenessProbe.enabled }}
livenessProbe:
  httpGet:
    path: {{ .Values.livenessProbe.httpGet.path }}
    port: {{ .Values.livenessProbe.httpGet.port }}
  initialDelaySeconds: {{ .Values.livenessProbe.initialDelaySeconds }}
  periodSeconds: {{ .Values.livenessProbe.periodSeconds }}
{{- end }}
{{- end }}

{{- define "yardpass-backend.readiness" -}}
{{- if .Values.readinessProbe.enabled }}
readinessProbe:
  httpGet:
    path: {{ .Values.readinessProbe.httpGet.path }}
    port: {{ .Values.readinessProbe.httpGet.port }}
  initialDelaySeconds: {{ .Values.readinessProbe.initialDelaySeconds }}
  periodSeconds: {{ .Values.readinessProbe.periodSeconds }}
{{- end }}
{{- end }}
