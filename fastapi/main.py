from fastapi import FastAPI, Header, HTTPException, Request, Query, UploadFile, File
from pydantic import BaseModel, Field
from typing import List, Optional, Dict, Any, Literal
import os, hmac, hashlib, time

# ------------------ Config ------------------
API_KEYS = set([k.strip() for k in os.getenv("ACP_API_KEYS", "").split(",") if k.strip()]) or {"test_key_123"}
WEBHOOK_SECRET = os.getenv("ACP_WEBHOOK_SECRET", "whsec_123")

# ------------------ App ------------------
app = FastAPI(title="ACP Gateway", version="0.1.0")

# ------------------ Storage (in-mem demo) ------------------
DB: Dict[str, Any] = {
    "products": [
        {
            "id": "sku_123",
            "title": "Work Boots",
            "description": "Durable work boots for construction and outdoor work",
            "variants": [{"id": "var_9", "size": "9", "price": {"value": 119.99, "currency": "USD"}}],
            "price": {"value": 119.99, "currency": "USD"},
            "available": True,
            "images": ["https://example.com/boots.jpg"],
            "tags": ["work", "boots", "durable"]
        },
        {
            "id": "sku_456",
            "title": "Safety Helmet",
            "description": "ANSI approved safety helmet",
            "variants": [{"id": "var_l", "size": "Large", "price": {"value": 45.99, "currency": "USD"}}],
            "price": {"value": 45.99, "currency": "USD"},
            "available": True,
            "images": ["https://example.com/helmet.jpg"],
            "tags": ["safety", "helmet", "ansi"]
        }
    ],
    "carts": {},
    "orders": {},
    "merchants": {},
    "audit": [],
}

# ------------------ Models ------------------
class Money(BaseModel):
    value: float
    currency: str = "USD"

class CartItem(BaseModel):
    product_id: str
    variant_id: Optional[str] = None
    quantity: int = 1
    price: Optional[Money] = None

class IntentPayload(BaseModel):
    action: str
    items: List[CartItem] = []
    notes: Optional[str] = None
    context: Optional[Dict[str, Any]] = None

class Intent(BaseModel):
    type: Literal["acp.intent"] = "acp.intent"
    actor: Literal["llm"] = "llm"
    payload: IntentPayload

class CheckoutReq(BaseModel):
    cart_id: str
    email: Optional[str] = None
    shipping_address: Optional[Dict[str, Any]] = None
    payment: Optional[Dict[str, Any]] = None

class MerchantRegister(BaseModel):
    name: str
    domain: str
    api_base: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None

# ------------------ Helpers ------------------

def require_api_key(auth: Optional[str]):
    if not auth or not auth.startswith("Bearer "):
        raise HTTPException(401, detail="missing bearer token")
    token = auth.split(" ", 1)[1]
    if token not in API_KEYS:
        raise HTTPException(401, detail="invalid api key")

def is_sandbox(req: Request, sandbox_header: Optional[str], sandbox_q: Optional[int]):
    if sandbox_header and sandbox_header.lower() == "true":
        return True
    if sandbox_q == 1:
        return True
    return False

def audit(event: str, data: Dict[str, Any]):
    DB["audit"].append({
        "ts": int(time.time()),
        "event": event,
        "data": data,
    })

def verify_hmac(raw_body: bytes, signature: str):
    # signature: "t=<ts>,v1=<hex>"
    try:
        parts = dict(p.split("=") for p in signature.split(","))
        v1 = parts.get("v1")
    except Exception:
        raise HTTPException(400, detail="bad signature header")
    mac = hmac.new(WEBHOOK_SECRET.encode(), raw_body, hashlib.sha256).hexdigest()
    if not hmac.compare_digest(mac, v1 or ""):
        raise HTTPException(401, detail="invalid signature")

# ------------------ Routes ------------------
@app.get("/healthz")
def healthz():
    return {"ok": True}

@app.get("/acp/v1/products")
def list_products(q: Optional[str] = Query(None), limit: int = 20, offset: int = 0, 
                  authorization: Optional[str] = Header(None)):
    require_api_key(authorization)
    items = DB["products"]
    if q:
        ql = q.lower()
        items = [p for p in items if ql in p["title"].lower() or ql in p["id"].lower() or ql in p.get("description", "").lower()]
    return {"ok": True, "items": items[offset: offset+limit], "total": len(items)}

@app.post("/acp/v1/intent")
async def handle_intent(intent: Intent, request: Request,
                        authorization: Optional[str] = Header(None),
                        x_acp_sandbox: Optional[str] = Header(None),
                        sandbox: Optional[int] = Query(None)):
    require_api_key(authorization)
    sb = is_sandbox(request, x_acp_sandbox, sandbox)

    # init cart
    cart_id = intent.payload.context.get("session_id") if intent.payload.context else None
    cart_id = cart_id or f"cart_{int(time.time()*1000)}"
    cart = DB["carts"].setdefault(cart_id, {"id": cart_id, "items": []})

    if intent.payload.action == "add_to_cart":
        for it in intent.payload.items:
            # naive price lookup
            prod = next((p for p in DB["products"] if p["id"] == it.product_id), None)
            if not prod:
                raise HTTPException(404, detail=f"product {it.product_id} not found")
            price = it.price or prod.get("price")
            cart["items"].append({
                "product_id": it.product_id,
                "variant_id": it.variant_id,
                "quantity": it.quantity,
                "price": price,
            })
    elif intent.payload.action == "select_variant":
        # Update variant for existing items
        for it in intent.payload.items:
            for cart_item in cart["items"]:
                if cart_item["product_id"] == it.product_id:
                    cart_item["variant_id"] = it.variant_id
    elif intent.payload.action == "set_quantity":
        # Update quantity for existing items
        for it in intent.payload.items:
            for cart_item in cart["items"]:
                if cart_item["product_id"] == it.product_id:
                    cart_item["quantity"] = it.quantity
    elif intent.payload.action == "checkout":
        # defer to /acp/v1/checkout
        pass
    else:
        raise HTTPException(400, detail="unknown action")

    subtotal = sum((i["price"]["value"] * i.get("quantity", 1)) for i in cart["items"])
    result = {
        "type": "acp.intent.result",
        "ok": True,
        "sandbox": sb,
        "cart": {**cart, "subtotal": {"value": round(subtotal, 2), "currency": "USD"}},
        "next": {"action": "checkout", "endpoint": "/acp/v1/checkout"},
    }
    audit("intent", {"sandbox": sb, "intent": intent.dict(), "result": result})
    return result

@app.post("/acp/v1/checkout")
async def checkout(req: CheckoutReq, request: Request,
                   authorization: Optional[str] = Header(None),
                   x_acp_sandbox: Optional[str] = Header(None),
                   sandbox: Optional[int] = Query(None)):
    require_api_key(authorization)
    sb = is_sandbox(request, x_acp_sandbox, sandbox)
    cart = DB["carts"].get(req.cart_id)
    if not cart:
        raise HTTPException(404, detail="cart not found")

    order_id = f"ord_{int(time.time()*1000)}"
    DB["orders"][order_id] = {
        "id": order_id,
        "cart_id": req.cart_id,
        "email": req.email,
        "items": cart["items"],
        "sandbox": sb,
        "status": "created",
        "shipping_address": req.shipping_address,
        "payment": req.payment,
    }
    resp = {
        "ok": True,
        "sandbox": sb,
        "order_id": order_id,
        "payment_intent_status": "requires_action" if sb else "succeeded",
        "redirect_url": f"https://merchant.example/checkout/{order_id}",
    }
    audit("checkout", {"sandbox": sb, "req": req.dict(), "resp": resp})
    return resp

@app.post("/merchant/register")
async def merchant_register(body: MerchantRegister, authorization: Optional[str] = Header(None)):
    require_api_key(authorization)
    mid = body.domain
    DB["merchants"][mid] = {"id": mid, **body.dict()}
    audit("merchant.register", body.dict())
    return {"ok": True, "merchant_id": mid}

@app.post("/merchant/feed")
async def merchant_feed(url: Optional[str] = Query(None), file: Optional[UploadFile] = File(None),
                        authorization: Optional[str] = Header(None)):
    require_api_key(authorization)
    # Demo: wipe + seed with one product or parse uploaded file
    DB["products"] = DB["products"][:1]
    if file:
        content = (await file.read()).decode("utf-8")[:5000]  # truncate demo
        audit("merchant.feed.upload", {"size": len(content)})
    elif url:
        audit("merchant.feed.url", {"url": url})
    return {"ok": True, "ingested": len(DB["products"]) }

@app.post("/webhooks/acp")
async def webhooks(request: Request, x_acp_signature: Optional[str] = Header(None)):
    raw = await request.body()
    if not x_acp_signature:
        raise HTTPException(401, detail="missing signature")
    verify_hmac(raw, x_acp_signature)
    payload = await request.json()
    audit("webhook", payload)
    return {"ok": True}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
