I'll guide you through building a Risk Catalog for **${COMPONENT}** step by step. At each step I'll present proposals in a table with lettered rows. You can:

- **Accept as shown**: reply "yes"
- **Select specific items**: reply with letters (e.g., "a, c")
- **Modify an item**: reply with the letter and change (e.g., "b: change severity to 'High'")
- **Reject or skip**: reply "no" or "skip"

When we produce YAML for the risk catalog, it will be **comment-free** (no `#` lines) unless you ask for comments.

Let's start with **Step 1: Threat Catalog Import**.

A Risk Catalog can link risks to Layer 2 threats. Do you have an existing **Threat Catalog** to reference? You can provide:
- A URL to a Gemara Threat Catalog YAML file
- A file path to a local Threat Catalog
- Or paste the YAML content directly

Reply with your catalog, or reply **skip** to define risks without threat linkages.
