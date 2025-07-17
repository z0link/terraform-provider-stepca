# Terraform Provider StepCA: Ideen fuer Datenquellen und Ressourcen

Diese Datei sammelt moegliche Terraform **Data Sources** und **Resources** fuer einen Step-CA Provider. Grundlage sind die Funktionen aus der Step-CA Dokumentation (z.B. `provisioners.mdx`, `templates.mdx`, `policies.mdx`, `configuration.mdx`) und die gaengigen Anforderungen an Terraform-Ressourcen:

* Es muss eine API oder CLI geben, die Erzeugung, Aenderung und Loeschung erlaubt.
* Die Operationen muessen idempotent und deterministisch sein, damit der Terraform-Status korrekt abgebildet wird.
* Kurzlebige oder lediglich lokale Artefakte eignen sich eher nicht als verwaltbare Ressourcen.

## Potenzielle Data Sources

- **stepca_version** – Liefert die Version des CA-Servers (bereits implementiert).
- **stepca_ca_certificate** – Gibt das Root- bzw. Intermediate-Zertifikat zurueck.
- **stepca_defaults** – Liest Einstellungen aus einer `defaults.json`.
- **stepca_template** – Gibt eine bestehende Zertifikatsvorlage zurueck.
- **stepca_provisioners** – Liefert eine Liste der konfigurierten Provisioner.
- **stepca_policy** – Gibt die aktuelle Issuance Policy aus.
- **stepca_provisioner_token** – Erstellt Einmal-Tokens fuer bestimmte Provisioner (z.B. OIDC oder ACME) ohne diese als Resource zu verwalten.

## Potenzielle Resources

- **stepca_certificate** – Signiert ein CSR und liefert das Zertifikat (bereits vorhanden; Create-only).
- **stepca_provisioner** – Legt Provisioner an (JWK, OIDC, ACME, X5C, SSHPOP, Cloud usw.), aendert und entfernt sie.
- **stepca_template** – Erstellt bzw. aktualisiert Zertifikats-Templates aus den Vorlagen der Step-CA Dokumentation.
- **stepca_policy** – Definiert Issuance Policies, wie in `policies.mdx` beschrieben.
- **stepca_webhook** – Verwaltung der Webhook-Konfiguration fuer Ereignisse (siehe `webhooks.mdx`).
- **stepca_ra_config** – Aktiviert und konfiguriert RA-Mode bzw. Remote Authorities.
- **stepca_ca_config** – Generiert `ca.json` bzw. steuert einzelne Felder wie Adressen, DB-Einstellungen oder SSH-Optionen.
- **stepca_cert_renewal** – Erneuert Zertifikate ueber das `/renew`-Endpoint (eher data source, da kurzlebig).
- **stepca_revocation** – Widerruft Zertifikate ueber `/revoke` (nur sinnvoll, wenn der Status nachvollziehbar ist).

Diese Liste ist nicht abschliessend und dient als Ausgangspunkt fuer kuenftige Erweiterungen des Providers.
