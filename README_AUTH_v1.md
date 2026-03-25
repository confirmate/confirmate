# Design Document: Hybrid Resource Authorization

## 1. Goal
To implement a scalable authorization system that manages access to various system resources (e.g., **Targets of Evaluation (ToE)**, **Catalogs**, and **AuditScopes**) we are using a tiered approach.

## 2. Authorization Hierarchy

The system separates **Identity** from **Resource Ownership** to balance global management with fine-grained access control.

### Tier 1: Global System Roles (via JWT)
- **Source:** Managed in the Identity Provider (e.g., Keycloak) via `resource_access` claims for a component specific admin role. And as `realm_access` claim for a system-wide admin role.
- **Function:** `resource_access` defines the user's functional role within the component (e.g., Orchestrator) and `realm_access` within the entire system (Confirmate).
- **Enforcement**: Handled by the **AuthInterceptor**
- **Logic:** If the JWT contains a global `admin` (`resource_access` or `realm_access`) role, the component grants access to **all** resources, bypassing local database checks.

### Tier 2: Scoped Resource Roles (via User DB)
- **Source:** Managed locally in the Orchestrators users database.
- **Function:** Defines the specific relationship between a **User** and a **Resource**.
- **Roles:**
    - **`OWNER`**: Full administrative control over the specific resource (Create, Delete, Manage Permissions).
    - **`CONTRIBUTOR`**: Write access. Can modify the resource but cannot delete it or manage permissions.
    - **`READER`**: Read-only access to the resource.

## 3. Database Schema (Specific Join Tables)

We use dedicated **Join Tables** (Many-to-Many) for each resource type.

### Entity Tables (Core Data)
- `users`: Local cache of user identities (ID from `sub` claim, email, etc.).
- Already existing:
  - `targets_of_evaluation`: Main table for ToE data.
  - `catalogs`: Main table for Catalog data.
  - `audit_scopes`: Main table for Scope data.
  - ...

### Permission Join Tables
Each table links a `user_id` to a specific `resource_id` with a designated `role`.

| Table Name | Columns | Constraints |
| :--- | :--- | :--- |
| **`toe_permissions`** | `user_id`, `toe_id`, `role` | PK(user_id, toe_id), FK(toe_id) **ON DELETE CASCADE** |
| **`catalog_permissions`** | `user_id`, `catalog_id`, `role` | PK(user_id, catalog_id), FK(catalog_id) **ON DELETE CASCADE** |
| **`audit_scope_permissions`** | `user_id`, `audit_scope_id`, `role` | PK(user_id, audit_scope_id), FK(scope_id) **ON DELETE CASCADE** |

### Decision for Join Tables
- Adding a new resource type only requires a new join table without touching existing logic.
- Permission changes in the database take effect immediately (no waiting for JWT expiration).
- Supports ON DELETE CASCADE. If a resource is deleted, its permissions are wiped automatically by the DB.
- Databases handle specific, indexed join tables much faster than one giant "catch-all" table.

## 4. User Lifecycle: Just-In-Time (JIT) Provisioning

To avoid manual synchronization, users are registered "on the fly" (JIT Provisioning) when they interact with the service for the first time.

When a user performs an action (like creating a resource), the system ensures the user exists locally:
1. **Extract Identity:** Get `user_id` (`sub` claim) from the validated JWT.
2. **Upsert User:** Perform an `UPSERT` operation on the `users` table. If the user doesn't exist, create them; if they do, update their metadata (e.g., last login).

## 5. Access Decision Logic

When an API request is made for a specific resource (e.g., `UpdateTargetOfEvaluation(id="XXX")`), the logic differs based on whether a user is **creating** a new resource or **accessing** an existing one.

### 5.1: Authentication & Global Bypass (Interceptor)
- **Check 1:** Is the JWT valid?
- **Check 2:** Does the user have the global `admin` role? 
  - **Yes:** **ALLOW** (Access granted immediately for support/maintenance).
  - **No:** Proceed to Step 5.2.

### 5.2 Creation Check (Global Check)
Before a resource is created, the system checks the **JWT**:
- **Requirement:** Does the token contain a specific `user` role? -> - **Yes:** Perform **JIT Provisioning** and create resource + ownership link. (Upsert)

### 5.3 Resource-Specific Check
For actions on existing resources (Read, Update, Delete) check 
**Local Permission (DB):** Check the specific join table for `user_id` + `resource_id`.
    - If a valid `role` (OWNER/EDITOR/VIEWER) is found → **ALLOW**.
    - Otherwise **DENY**

