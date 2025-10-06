
# How-To: Deploy a Secure Go Application on Oracle Cloud with K3s, Traefik, and Cert-Manager

This guide provides step-by-step instructions to set up a production-ready Kubernetes environment on Oracle Cloud Infrastructure (OCI) using K3s, Traefik, and Cert-Manager. It covers networking, cluster installation, secrets management, and automated HTTPS. **All steps should be run on the remote server via SSH unless specified otherwise.**

---

## Step 1: OCI Instance and Firewall Setup

1. **Create an OCI Compute Instance:**
   - Use an "Always Free Eligible" shape (e.g., `VM.Standard.E2.1.Micro`).
   - Select the Canonical Ubuntu image.
   - Ensure it is assigned a Public IPv4 address.
   - Add your public SSH key for access.

2. **Configure VCN Security Lists:**
   - Go to your instance's Virtual Cloud Network (VCN) → Security Lists → Default Security List.
   - Add the following Ingress Rules:
     - **Port 22 (SSH):** Source `0.0.0.0/0`, Protocol TCP, Destination Port 22
     - **Port 80 (HTTP):** Source `0.0.0.0/0`, Protocol TCP, Destination Port 80 _(required for Let's Encrypt validation)_
     - **Port 443 (HTTPS):** Source `0.0.0.0/0`, Protocol TCP, Destination Port 443
     - **Port 6443 (K8s API):** Source `0.0.0.0/0`, Protocol TCP, Destination Port 6443

---

## Step 2: Install and Configure K3s and Local Tooling

1. **SSH into the OCI Instance.**

2. **Install K3s (Kubernetes):**
   - Disable the default Traefik Ingress controller (we will install our own):
     ```sh
     curl -sfL https://get.k3s.io | sh -s - --disable=traefik
     ```

3. **Configure kubectl for Server-Side Use:**
   - K3s includes `kubectl`, but you need to configure your shell to find it:
     ```sh
     mkdir -p ~/.kube
     sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
     sudo chown $(id -u):$(id -g) ~/.kube/config
     ```
   - Your `kubectl` commands will now work correctly on the server.

4. **Install Helm (Kubernetes Package Manager):**
   ```sh
   curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
   chmod 700 get_helm.sh
   ./get_helm.sh
   ```

5. **Clone Your Application Repository:**
   ```sh
   sudo apt update && sudo apt install git -y
   git clone https://github.com/CharlRitter/brewsource-mcp.git
   cd brewsource-mcp
   ```

---

## Step 3: Install Cluster Controllers

1. **Install Traefik (Ingress Controller):**
   - Uninstall any old versions first:
     ```sh
     helm uninstall traefik
     ```
   - Install Traefik with explicit arguments:
     ```sh
     helm repo add traefik https://helm.traefik.io/traefik
     helm repo update
     helm install traefik traefik/traefik \
       --set="additionalArguments={--providers.kubernetesingress.ingressclass=traefik}" \
       --set="ports.websecure.tls.enabled=true"
     ```

2. **Install Cert-Manager (for SSL/TLS):**
   ```sh
   helm repo add jetstack https://charts.jetstack.io
   helm repo update
   helm install cert-manager jetstack/cert-manager \
     --namespace cert-manager \
     --create-namespace \
     --set installCRDs=true
   ```

3. **Install Sealed Secrets (for Encrypted Secrets):**
   ```sh
   helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
   helm install sealed-secrets sealed-secrets/sealed-secrets -n kube-system
   ```

_Wait about a minute for all controller pods to be in a Running state before proceeding._

---

## Step 4: Nuke, Pave, and Deploy the Application

This "nuke and pave" approach ensures both the application and the database start with the same, correct credentials, preventing authentication failures.

1. **Delete the Namespace (if it exists):**
   ```sh
   kubectl delete namespace brewsource-prod
   ```

2. **Create the Namespace:**
   ```sh
   kubectl create namespace brewsource-prod
   ```

3. **Create and Seal the Database Secret:**
   - **IMPORTANT:** The password must be URL-safe. Use only letters (a-z, A-Z) and numbers (0-9). Special characters will cause the application to crash.

   - Create a temporary plain-text secret file:
     ```sh
     cat <<EOF > temp-postgres-secret.yaml
     apiVersion: v1
     kind: Secret
     metadata:
       name: postgres-secret
       namespace: brewsource-prod
     stringData:
       POSTGRES_USER: "brewsource_user"
       POSTGRES_PASSWORD: "averylongsafepassword123"
     EOF
     ```

   - Find the exact name of the Sealed Secrets service:
     ```sh
     kubectl get svc -n kube-system
     # The name will be 'sealed-secrets'
     ```

   - Use `kubeseal` to encrypt the file, specifying the correct controller name and namespace. This overwrites the old `postgres-sealedsecret.yaml` in your repo:
     ```sh
     kubeseal --controller-namespace kube-system --controller-name sealed-secrets < temp-postgres-secret.yaml > k8s/prod/postgres-sealedsecret.yaml
     ```

   - Clean up the temporary plain-text file:
     ```sh
     rm temp-postgres-secret.yaml
     ```

4. **Deploy the Application Stack:**
   - Apply your entire Kustomize configuration. This will create all resources, including the ClusterIssuer for cert-manager and the final, corrected Ingress rules:
     ```sh
     kubectl apply -k brewsource-mcp/k8s/prod
     ```

---

## Step 5: Final DNS Configuration

1. **Find your server's public IP address.**

2. **Go to your DNS provider** (the service where you manage your domain).

3. **Create an A Record:**
   - **Type:** A
   - **Host/Name:** `brewsource` (or your chosen subdomain)
   - **Value:** Your server's public IP address

_After DNS propagates (a few minutes to an hour), cert-manager will automatically complete the SSL/TLS challenge, and your site will be live and secure._

## Step 6: Run Codacy CLI Analysis

1. **Install Codacy CLI:**
  - Follow the instructions on the [Codacy CLI documentation](https://docs.codacy.com/cli/installation/) to install the CLI tool.

2. **Run the analysis:**
  - Navigate to your project directory and run:
    ```sh
    codacy-cli analyze
    ```

3. **Review the results:**
  - After the analysis completes, review the results and make any necessary changes to improve code quality.
