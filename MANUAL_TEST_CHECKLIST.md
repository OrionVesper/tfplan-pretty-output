# 🧪 tfplan-pretty-output — Manual Test Checklist

Use this checklist after any significant change, or as a periodic regression check.
Total time: ~20-30 minutes for full pass, ~5 minutes for smoke subset (marked ⚡).

---

## 🛠️ Setup

Before testing, prepare a clean sandbox:

```bash
# 1. Wipe storage to start fresh
rm -rf ~/.tfplan-pretty-output/

# 2. Create a sandbox terraform project
mkdir -p /tmp/tfsandbox
cd /tmp/tfsandbox
cat > main.tf <<'EOF'
resource "null_resource" "demo" {
  triggers = {
    timestamp = timestamp()
  }
}
resource "null_resource" "demo2" {
  triggers = {
    name = "test"
  }
}
EOF

# 3. Initialize terraform
terraform init