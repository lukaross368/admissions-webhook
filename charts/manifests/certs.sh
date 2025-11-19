openssl genrsa -out tls.key 2048
openssl req -x509 -new -nodes -key tls.key -days 365 \
  -out tls.crt \
  -subj "/CN=webhook.admissions-webhook.svc" \
  -addext "subjectAltName = DNS:webhook,DNS:webhook.admissions-webhook,DNS:webhook.admissions-webhook.svc,DNS:webhook.admissions-webhook.svc.cluster.local"