import json
from pathlib import Path


ROOT = Path(r"c:\Users\kaush\Desktop\Healthonyx\docs")
SPRINT3 = ROOT / "Healthonyx-API-Sprint3-Implemented.postman_collection.json"
REGRESSION = ROOT / "Healthonyx-API-Complete-Regression.postman_collection.json"


def walk_items(items):
    for item in items:
        yield item
        if "item" in item:
            yield from walk_items(item["item"])


def find_item(items, name):
    for item in walk_items(items):
        if item.get("name") == name:
            return item
    return None


def add_test_event(item, lines):
    item["event"] = [
        {
            "listen": "test",
            "script": {"type": "text/javascript", "exec": lines},
        }
    ]


def ensure_variable(collection, key, value=""):
    vars_ = collection.setdefault("variable", [])
    for v in vars_:
        if v.get("key") == key:
            if value != "" and not v.get("value"):
                v["value"] = value
            return
    vars_.append({"key": key, "value": value})


def remove_request_by_name(items, name):
    i = 0
    while i < len(items):
        it = items[i]
        if it.get("name") == name and "request" in it:
            items.pop(i)
            continue
        if "item" in it:
            remove_request_by_name(it["item"], name)
        i += 1


def add_expected_failure(items, folder_name, req_name, method, raw_url, expected_code):
    folder = find_item(items, folder_name)
    if not folder or "item" not in folder:
        return
    if find_item(folder["item"], req_name):
        return
    path_tokens = raw_url.replace("{{baseUrl}}", "").lstrip("/").split("/")
    req = {
        "name": req_name,
        "event": [
            {
                "listen": "test",
                "script": {
                    "type": "text/javascript",
                    "exec": [
                        f"pm.test('status is {expected_code}', function () {{",
                        f"  pm.response.to.have.status({expected_code});",
                        "});",
                    ],
                },
            }
        ],
        "request": {
            "method": method,
            "header": [{"key": "Authorization", "value": "Bearer {{patientToken}}"}],
            "url": {"raw": raw_url, "host": ["{{baseUrl}}"], "path": path_tokens},
        },
    }
    folder["item"].append(req)


def harden_sprint3():
    data = json.loads(SPRINT3.read_text(encoding="utf-8"))
    items = data.get("item", [])

    ensure_variable(data, "hospitalId", "")
    ensure_variable(data, "targetUserId", "")

    hospitals = find_item(items, "GET /api/hospitals/near")
    if hospitals:
        add_test_event(
            hospitals,
            [
                "pm.test('status is 200', function () { pm.response.to.have.status(200); });",
                "const j = pm.response.json();",
                "pm.test('hospitals is array', function () { pm.expect(j.hospitals).to.be.an('array'); });",
                "if (j.hospitals && j.hospitals.length > 0) {",
                "  pm.collectionVariables.set('hospitalId', j.hospitals[0].id);",
                "}",
            ],
        )

    admin_users = find_item(items, "GET /api/admin/users")
    if admin_users:
        add_test_event(
            admin_users,
            [
                "pm.test('status is 200', function () { pm.response.to.have.status(200); });",
                "const j = pm.response.json();",
                "pm.test('users is array', function () { pm.expect(j.users).to.be.an('array'); });",
                "const nonAdmin = (j.users || []).find(u => u.role !== 'admin');",
                "if (nonAdmin && nonAdmin.id) {",
                "  pm.collectionVariables.set('targetUserId', nonAdmin.id);",
                "}",
            ],
        )

    toggle_user = find_item(items, "PATCH /api/admin/users/:id (toggle active)")
    if toggle_user:
        add_test_event(
            toggle_user,
            [
                "pm.test('status is 200', function () { pm.response.to.have.status(200); });",
            ],
        )

    req_reschedule = find_item(items, "PATCH /api/appointments/:id/request-reschedule (patient)")
    if req_reschedule:
        add_test_event(
            req_reschedule,
            [
                "pm.test('status is expected business failure 400', function () {",
                "  pm.response.to.have.status(400);",
                "});",
            ],
        )

    add_expected_failure(
        items,
        "Geo and Doctor Discovery",
        "GET /api/hospitals/:id/departments (invalid id -> 400)",
        "GET",
        "{{baseUrl}}/api/hospitals/not-a-uuid/departments",
        400,
    )

    SPRINT3.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")


def harden_regression():
    data = json.loads(REGRESSION.read_text(encoding="utf-8"))
    items = data.get("item", [])

    ensure_variable(data, "hospitalId", "")

    remove_request_by_name(items, "GET /api/doctor/slots")

    hospitals = find_item(items, "GET /api/hospitals/near")
    if hospitals:
        add_test_event(
            hospitals,
            [
                "pm.test('status is 200', function () { pm.response.to.have.status(200); });",
                "const j = pm.response.json();",
                "pm.test('hospitals is array', function () { pm.expect(j.hospitals).to.be.an('array'); });",
                "if (j.hospitals && j.hospitals.length > 0) {",
                "  pm.collectionVariables.set('hospitalId', j.hospitals[0].id);",
                "}",
            ],
        )

    departments = find_item(items, "GET /api/hospitals/:id/departments")
    if departments:
        add_test_event(
            departments,
            [
                "pm.test('status is 200', function () { pm.response.to.have.status(200); });",
                "const j = pm.response.json();",
                "pm.test('departments is array', function () { pm.expect(j.departments).to.be.an('array'); });",
            ],
        )

    add_expected_failure(
        items,
        "Sprint 3.5",
        "GET /api/hospitals/:id/departments (invalid id -> 400)",
        "GET",
        "{{baseUrl}}/api/hospitals/not-a-uuid/departments",
        400,
    )

    REGRESSION.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    harden_sprint3()
    harden_regression()
    print("Hardened Postman collections.")
