#!/usr/bin/env python3
"""Apply implementation seed data to ControlInScope records via REST.

Usage:
    apply-implementation-seed.py <seed_file> <controls.json> <cis.json> <users.json> <token_file> <api_url_file>
"""
import json
import sys
import urllib.request


def main():
    seed_file = sys.argv[1]
    controls_file = sys.argv[2]
    cis_file = sys.argv[3]
    users_file = sys.argv[4]
    token_file = sys.argv[5]
    api_url_file = sys.argv[6]

    with open(seed_file) as f:
        seed = json.load(f)
    with open(controls_file) as f:
        controls_json = json.load(f)
    with open(cis_file) as f:
        cis_json = json.load(f)
    with open(users_file) as f:
        users_json = json.load(f)
    with open(token_file) as f:
        token = f.read().strip()
    with open(api_url_file) as f:
        api = f.read().strip()

    # Build shortName -> controlId (including sub-controls)
    short_to_ctrl = {}

    def walk(c):
        short_to_ctrl[c.get("shortName", "")] = c.get("id", "")
        for sub in c.get("controls", []) or []:
            walk(sub)

    for c in controls_json.get("controls", []):
        walk(c)

    # Build controlId -> controlInScopeId
    ctrl_to_cis = {}
    for cis in cis_json.get("controlsInScope", []):
        ctrl_to_cis[cis.get("controlId", "")] = cis.get("id", "")

    # Build username -> userId
    username_to_id = {}
    for u in users_json.get("users", []):
        username_to_id[u.get("username", "")] = u.get("id", "")

    def api_call(method, path, body=None):
        url = api + path
        data = json.dumps(body).encode() if body else None
        req = urllib.request.Request(url, data=data, method=method)
        req.add_header("Authorization", "Bearer " + token)
        req.add_header("Content-Type", "application/json")
        try:
            with urllib.request.urlopen(req) as resp:
                return json.loads(resp.read()) if resp.status == 200 else None
        except Exception:
            return None

    # State machine: OPEN -> IN_PROGRESS -> IMPLEMENTED -> READY_FOR_REVIEW -> ACCEPTED
    state_order = [
        "CONTROL_IN_SCOPE_STATE_IN_PROGRESS",
        "CONTROL_IN_SCOPE_STATE_IMPLEMENTED",
        "CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW",
        "CONTROL_IN_SCOPE_STATE_ACCEPTED",
    ]

    for entry in seed.get("controls", []):
        sn = entry.get("shortName", "")
        state = entry.get("state", "")
        details = entry.get("details", "")
        assignee = entry.get("assignee", "")

        ctrl_id = short_to_ctrl.get(sn)
        if not ctrl_id:
            print(f"  WARNING: control {sn} not found in catalog")
            continue

        cis_id = ctrl_to_cis.get(ctrl_id)
        if not cis_id:
            print(f"  WARNING: controlInScope for {sn} not found")
            continue

        assignee_id = username_to_id.get(assignee, "")

        # Update implementation details and assignee
        body = {"id": cis_id}
        if details:
            body["implementationDetails"] = details
        if assignee_id:
            body["assigneeId"] = assignee_id
        api_call("PUT", f"/v1/orchestrator/controls_in_scope/{cis_id}", body)

        # Transition state, following the state machine step by step
        transition_comments = {
            "CONTROL_IN_SCOPE_STATE_IN_PROGRESS": "Implementation work started",
            "CONTROL_IN_SCOPE_STATE_IMPLEMENTED": "Implementation completed, ready for review",
            "CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW": "Submitted for security review",
            "CONTROL_IN_SCOPE_STATE_ACCEPTED": "Reviewed and accepted by security team",
        }
        if state in state_order:
            target_idx = state_order.index(state)
            for i in range(target_idx + 1):
                api_call(
                    "POST",
                    f"/v1/orchestrator/controls_in_scope/{cis_id}/transition",
                    {"id": cis_id, "toState": state_order[i], "comment": transition_comments.get(state_order[i], "State transition")},
                )

        label = state.replace("CONTROL_IN_SCOPE_STATE_", "")
        print(f"  Set {sn} -> {label} (assignee: {assignee})")


if __name__ == "__main__":
    main()
