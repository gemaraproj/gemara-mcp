I'll guide you through building a Policy for **${COMPONENT}** step by step. At each step I'll present proposals in a table with lettered rows. You can:

- **Accept as shown**: reply "yes"
- **Select specific items**: reply with letters (e.g., "a, c")
- **Modify an item**: reply with the letter and change (e.g., "b: update justification to 'residual risk within appetite'")
- **Reject or skip**: reply "no" or "skip"

When we produce YAML for the policy, it will be **comment-free** (no `#` lines) unless you ask for comments.

Let's start with **Step 1: Catalog and Artifact Import**.

A Policy imports Control Catalogs (Layer 2) to define which security controls apply, and optionally imports Guidance Catalogs (Layer 1), Risk Catalogs (Layer 3), and other Policies (Layer 3).

Do you have an existing **Control Catalog** to import? You can provide:
- A URL to a Gemara Control Catalog YAML file
- A file path to a local Control Catalog
- Or paste the YAML content directly

Reply with your catalog, or reply **skip** to define imports later.
