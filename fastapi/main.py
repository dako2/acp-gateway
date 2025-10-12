from fastapi import FastAPI, Header, HTTPException, Request, Query, UploadFile, File
from pydantic import BaseModel, Field
from typing import List, Optional, Dict, Any, Literal
import os, hmac, hashlib, time
from datetime import datetime, timedelta
try:
    # Try relative imports first (when running as module)
    from .product_feed_models import (
        ProductFeed, FeedIngestionRequest, FeedIngestionResponse, 
        FeedValidationResult, EnhancedProduct, ProductSearchRequest, ProductSearchResponse,
        example_product_feed
    )
    from .product_feed_validation import (
        ProductFeedValidator, ProductFeedProcessor, ProductFeedTransformer, FeedStorage
    )
    from .agentic_checkout_models import (
        CheckoutSessionCreateRequest, CheckoutSessionUpdateRequest, CheckoutSessionCompleteRequest,
        CheckoutSession, CheckoutSessionWithOrder, Order, WebhookEvent, EventData,
        example_checkout_session, example_checkout_session_create_request
    )
except ImportError:
    # Fall back to absolute imports (when running directly)
    from product_feed_models import (
        ProductFeed, FeedIngestionRequest, FeedIngestionResponse, 
        FeedValidationResult, EnhancedProduct, ProductSearchRequest, ProductSearchResponse,
        example_product_feed
    )
    from product_feed_validation import (
        ProductFeedValidator, ProductFeedProcessor, ProductFeedTransformer, FeedStorage
    )
    from agentic_checkout_models import (
        CheckoutSessionCreateRequest, CheckoutSessionUpdateRequest, CheckoutSessionCompleteRequest,
        CheckoutSession, CheckoutSessionWithOrder, Order, WebhookEvent, EventData,
        example_checkout_session, example_checkout_session_create_request
    )

# ------------------ Config ------------------
API_KEYS = set([k.strip() for k in os.getenv("ACP_API_KEYS", "").split(",") if k.strip()]) or {"test_key_123"}
WEBHOOK_SECRET = os.getenv("ACP_WEBHOOK_SECRET", "whsec_123")

# ------------------ App ------------------
app = FastAPI(
    title="ACP Gateway",
    version="0.1.0",
    description="Agentic Commerce Protocol (ACP) Gateway - Reference Implementation",
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json"
)

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
    "enhanced_products": {},  # Store for OpenAI Product Feed products
    "carts": {},
    "checkout_sessions": {},  # Store for Agentic Checkout sessions
    "orders": {},
    "merchants": {},
    "audit": [],
}

# ------------------ Product Feed Components ------------------
feed_validator = ProductFeedValidator()
feed_processor = ProductFeedProcessor()
feed_transformer = ProductFeedTransformer()
feed_storage = FeedStorage()

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

# ------------------ OpenAI Product Feed Endpoints ------------------

@app.post("/merchant/feed/ingest")
async def ingest_product_feed(
    request: FeedIngestionRequest,
    authorization: Optional[str] = Header(None)
):
    """Ingest a product feed according to OpenAI Product Feed Specification"""
    require_api_key(authorization)
    
    # Validate the feed
    validation_result = feed_validator.validate_feed_ingestion_request(request)
    
    if not validation_result.valid:
        return FeedIngestionResponse(
            ok=False,
            processed_count=0,
            error_count=len(validation_result.errors or []),
            errors=[f"{err.field}: {err.message}" for err in validation_result.errors or []],
            message="Feed validation failed"
        )
    
    # Process the feed
    processed_count = 0
    error_count = 0
    errors = []
    
    for product_feed in request.feed_data:
        try:
            # Convert to enhanced product
            enhanced_product = feed_transformer.product_feed_to_enhanced_product(
                product_feed, request.merchant_id
            )
            
            # Store in enhanced products
            DB["enhanced_products"][product_feed.basic_data.id] = enhanced_product.dict()
            
            # Also store in legacy format for backward compatibility
            legacy_product = feed_transformer.product_feed_to_legacy_product(product_feed)
            DB["products"].append(legacy_product)
            
            processed_count += 1
        except Exception as e:
            error_count += 1
            errors.append(f"Product {product_feed.basic_data.id}: {str(e)}")
    
    # Generate feed ID
    feed_id = f"feed_{int(time.time()*1000)}"
    
    # Store feed status
    feed_status = {
        "feed_id": feed_id,
        "merchant_id": request.merchant_id,
        "status": "completed" if error_count == 0 else "completed_with_errors",
        "processed_at": datetime.utcnow().isoformat(),
        "product_count": len(request.feed_data),
        "error_count": error_count,
        "last_updated": datetime.utcnow().isoformat()
    }
    feed_storage.store_feed(feed_id, feed_status)
    
    audit("merchant.feed.ingest", {
        "merchant_id": request.merchant_id,
        "feed_id": feed_id,
        "processed_count": processed_count,
        "error_count": error_count,
        "format": request.format
    })
    
    return FeedIngestionResponse(
        ok=error_count == 0,
        processed_count=processed_count,
        error_count=error_count,
        errors=errors if errors else None,
        feed_id=feed_id,
        message=f"Successfully processed {processed_count} products" if error_count == 0 else f"Processed {processed_count} products with {error_count} errors"
    )

@app.post("/merchant/feed/upload")
async def upload_product_feed(
    file: UploadFile = File(...),
    format_type: str = Query(..., description="Feed format (json, csv, tsv, xml)"),
    merchant_id: str = Query(..., description="Merchant ID"),
    authorization: Optional[str] = Header(None)
):
    """Upload and process a product feed file"""
    require_api_key(authorization)
    
    if format_type.lower() not in ["json", "csv", "tsv", "xml"]:
        raise HTTPException(400, detail="Invalid format type. Must be json, csv, tsv, or xml")
    
    try:
        # Read file content
        content = await file.read()
        
        # Process the feed
        feed_request = feed_processor.process_feed(format_type, content)
        feed_request.merchant_id = merchant_id
        
        # Validate and ingest
        validation_result = feed_validator.validate_feed_ingestion_request(feed_request)
        
        if not validation_result.valid:
            return FeedIngestionResponse(
                ok=False,
                processed_count=0,
                error_count=len(validation_result.errors or []),
                errors=[f"{err.field}: {err.message}" for err in validation_result.errors or []],
                message="Feed validation failed"
            )
        
        # Process each product
        processed_count = 0
        error_count = 0
        errors = []
        
        for product_feed in feed_request.feed_data:
            try:
                # Convert to enhanced product
                enhanced_product = feed_transformer.product_feed_to_enhanced_product(
                    product_feed, merchant_id
                )
                
                # Store in enhanced products
                DB["enhanced_products"][product_feed.basic_data.id] = enhanced_product.dict()
                
                # Also store in legacy format for backward compatibility
                legacy_product = feed_transformer.product_feed_to_legacy_product(product_feed)
                DB["products"].append(legacy_product)
                
                processed_count += 1
            except Exception as e:
                error_count += 1
                errors.append(f"Product {product_feed.basic_data.id}: {str(e)}")
        
        # Generate feed ID
        feed_id = f"feed_{int(time.time()*1000)}"
        
        audit("merchant.feed.upload", {
            "merchant_id": merchant_id,
            "feed_id": feed_id,
            "filename": file.filename,
            "format": format_type,
            "processed_count": processed_count,
            "error_count": error_count
        })
        
        return FeedIngestionResponse(
            ok=error_count == 0,
            processed_count=processed_count,
            error_count=error_count,
            errors=errors if errors else None,
            feed_id=feed_id,
            message=f"Successfully processed {processed_count} products from {file.filename}"
        )
        
    except Exception as e:
        raise HTTPException(400, detail=f"Failed to process feed: {str(e)}")

@app.post("/merchant/feed/validate")
async def validate_product_feed(
    feed: ProductFeed,
    authorization: Optional[str] = Header(None)
):
    """Validate a single product feed entry"""
    require_api_key(authorization)
    
    validation_result = feed_validator.validate_product_feed(feed)
    
    audit("merchant.feed.validate", {
        "product_id": feed.basic_data.id,
        "valid": validation_result.valid,
        "error_count": len(validation_result.errors or []),
        "warning_count": len(validation_result.warnings or [])
    })
    
    return validation_result

@app.get("/merchant/feed/status/{feed_id}")
async def get_feed_status(
    feed_id: str,
    authorization: Optional[str] = Header(None)
):
    """Get the status of a product feed"""
    require_api_key(authorization)
    
    feed_status = feed_storage.get_feed(feed_id)
    if not feed_status:
        raise HTTPException(404, detail="Feed not found")
    
    return feed_status

@app.get("/merchant/feeds")
async def list_feeds(
    merchant_id: Optional[str] = Query(None),
    authorization: Optional[str] = Header(None)
):
    """List all product feeds"""
    require_api_key(authorization)
    
    feeds = feed_storage.list_feeds()
    if merchant_id:
        feeds = [f for f in feeds if f.merchant_id == merchant_id]
    
    return {"feeds": feeds}

@app.post("/acp/v1/products/search")
async def search_products(
    search_request: ProductSearchRequest,
    authorization: Optional[str] = Header(None)
):
    """Search products using OpenAI Product Feed Specification fields"""
    require_api_key(authorization)
    
    # Get all enhanced products
    enhanced_products = list(DB["enhanced_products"].values())
    
    # Apply filters
    filtered_products = []
    for product_dict in enhanced_products:
        product_feed = ProductFeed(**product_dict["product_feed"])
        
        # Apply query filter
        if search_request.query:
            query_lower = search_request.query.lower()
            if not (query_lower in product_feed.basic_data.title.lower() or 
                   query_lower in product_feed.basic_data.description.lower()):
                continue
        
        # Apply category filter
        if search_request.category:
            if search_request.category.lower() not in product_feed.item_info.product_category.lower():
                continue
        
        # Apply brand filter
        if search_request.brand:
            if not product_feed.item_info.brand or search_request.brand.lower() not in product_feed.item_info.brand.lower():
                continue
        
        # Apply price filters
        if search_request.min_price:
            if product_feed.price_promotions.price.value < search_request.min_price.value:
                continue
        
        if search_request.max_price:
            if product_feed.price_promotions.price.value > search_request.max_price.value:
                continue
        
        # Apply availability filter
        if search_request.availability:
            if product_feed.availability_inventory.availability != search_request.availability:
                continue
        
        # Apply merchant filter
        if search_request.merchant_id:
            if product_dict["merchant_id"] != search_request.merchant_id:
                continue
        
        filtered_products.append(EnhancedProduct(**product_dict))
    
    # Apply sorting
    if search_request.sort_by:
        if search_request.sort_by == "price":
            filtered_products.sort(
                key=lambda p: p.product_feed.price_promotions.price.value,
                reverse=search_request.sort_order == "desc"
            )
        elif search_request.sort_by == "title":
            filtered_products.sort(
                key=lambda p: p.product_feed.basic_data.title,
                reverse=search_request.sort_order == "desc"
            )
    
    # Apply pagination
    total = len(filtered_products)
    offset = search_request.offset or 0
    limit = search_request.limit or 20
    paginated_products = filtered_products[offset:offset + limit]
    
    return ProductSearchResponse(
        products=paginated_products,
        total=total,
        limit=limit,
        offset=offset,
        has_more=offset + limit < total
    )

@app.get("/acp/v1/products/enhanced")
async def list_enhanced_products(
    q: Optional[str] = Query(None),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
    authorization: Optional[str] = Header(None)
):
    """List enhanced products with OpenAI Product Feed Specification compliance"""
    require_api_key(authorization)
    
    enhanced_products = list(DB["enhanced_products"].values())
    
    # Apply search filter
    if q:
        q_lower = q.lower()
        enhanced_products = [
            p for p in enhanced_products 
            if (q_lower in p["product_feed"]["basic_data"]["title"].lower() or 
                q_lower in p["product_feed"]["basic_data"]["description"].lower())
        ]
    
    # Apply pagination
    total = len(enhanced_products)
    paginated_products = enhanced_products[offset:offset + limit]
    
    return {
        "ok": True,
        "products": [EnhancedProduct(**p) for p in paginated_products],
        "total": total,
        "limit": limit,
        "offset": offset,
        "has_more": offset + limit < total
    }

# ==================== AGENTIC CHECKOUT ENDPOINTS ====================

@app.post("/checkout_sessions", response_model=CheckoutSession, status_code=201)
async def create_checkout_session(
    req: CheckoutSessionCreateRequest,
    request: Request,
    authorization: Optional[str] = Header(None),
    x_acp_sandbox: Optional[str] = Header(None),
    sandbox: Optional[int] = Query(None)
):
    """Create a new checkout session"""
    require_api_key(authorization)
    sb = is_sandbox(request, x_acp_sandbox, sandbox)
    
    # Generate checkout session ID
    session_id = f"checkout_session_{int(time.time() * 1000)}"
    
    # Calculate line items and totals
    line_items = []
    total_amount = 0
    
    for item in req.items:
        # Find product in our store
        product = next((p for p in DB["products"] if p["id"] == item.id), None)
        if not product:
            raise HTTPException(400, detail=f"Product {item.id} not found")
        
        # Calculate pricing
        base_amount = int(product["price"]["value"] * 100 * item.quantity)  # Convert to cents
        discount = 0
        subtotal = base_amount - discount
        tax = int(subtotal * 0.1)  # 10% tax
        total = subtotal + tax
        total_amount += total
        
        line_items.append({
            "id": f"line_item_{item.id}_{item.quantity}",
            "item": {"id": item.id, "quantity": item.quantity},
            "base_amount": base_amount,
            "discount": discount,
            "subtotal": subtotal,
            "tax": tax,
            "total": total
        })
    
    # Determine status based on fulfillment address
    status = "ready_for_payment" if req.fulfillment_address else "not_ready_for_payment"
    
    # Create fulfillment options if address provided
    fulfillment_options = []
    if req.fulfillment_address:
        fulfillment_options = [
            {
                "type": "shipping",
                "id": "standard_shipping",
                "title": "Standard Shipping",
                "subtitle": "Arrives in 4-5 business days",
                "carrier": "USPS",
                "earliest_delivery_time": (datetime.now() + timedelta(days=4)).isoformat(),
                "latest_delivery_time": (datetime.now() + timedelta(days=5)).isoformat(),
                "subtotal": 500,  # $5.00 in cents
                "tax": 50,
                "total": 550
            },
            {
                "type": "shipping",
                "id": "express_shipping",
                "title": "Express Shipping",
                "subtitle": "Arrives in 1-2 business days",
                "carrier": "FedEx",
                "earliest_delivery_time": (datetime.now() + timedelta(days=1)).isoformat(),
                "latest_delivery_time": (datetime.now() + timedelta(days=2)).isoformat(),
                "subtotal": 1500,  # $15.00 in cents
                "tax": 150,
                "total": 1650
            }
        ]
    
    # Calculate totals
    shipping_total = 550 if req.fulfillment_address else 0  # Default to standard shipping
    totals = [
        {"type": "items_base_amount", "display_text": "Item(s) total", "amount": total_amount},
        {"type": "subtotal", "display_text": "Subtotal", "amount": total_amount},
        {"type": "tax", "display_text": "Tax", "amount": int(total_amount * 0.1)},
    ]
    
    if req.fulfillment_address:
        totals.extend([
            {"type": "fulfillment", "display_text": "Shipping", "amount": shipping_total},
            {"type": "total", "display_text": "Total", "amount": total_amount + int(total_amount * 0.1) + shipping_total}
        ])
    else:
        totals.append({"type": "total", "display_text": "Total", "amount": total_amount + int(total_amount * 0.1)})
    
    # Create checkout session
    checkout_session = {
        "id": session_id,
        "buyer": req.buyer,
        "payment_provider": {"provider": "stripe", "supported_payment_methods": ["card"]},
        "status": status,
        "currency": "usd",
        "line_items": line_items,
        "fulfillment_address": req.fulfillment_address,
        "fulfillment_options": fulfillment_options,
        "fulfillment_option_id": "standard_shipping" if req.fulfillment_address else None,
        "totals": totals,
        "messages": [],
        "links": [
            {"type": "terms_of_use", "url": "https://example.com/terms"},
            {"type": "privacy_policy", "url": "https://example.com/privacy"}
        ]
    }
    
    # Store checkout session
    DB["checkout_sessions"][session_id] = checkout_session
    
    audit("checkout_session_created", {"session_id": session_id, "request": req.dict()})
    
    return checkout_session


@app.post("/checkout_sessions/{session_id}", response_model=CheckoutSession)
async def update_checkout_session(
    session_id: str,
    req: CheckoutSessionUpdateRequest,
    request: Request,
    authorization: Optional[str] = Header(None),
    x_acp_sandbox: Optional[str] = Header(None),
    sandbox: Optional[int] = Query(None)
):
    """Update an existing checkout session"""
    require_api_key(authorization)
    sb = is_sandbox(request, x_acp_sandbox, sandbox)
    
    # Find checkout session
    session = DB["checkout_sessions"].get(session_id)
    if not session:
        raise HTTPException(404, detail="Checkout session not found")
    
    # Update session with new data
    if req.buyer is not None:
        session["buyer"] = req.buyer
    
    if req.items is not None:
        # Recalculate line items
        line_items = []
        total_amount = 0
        
        for item in req.items:
            product = next((p for p in DB["products"] if p["id"] == item.id), None)
            if not product:
                raise HTTPException(400, detail=f"Product {item.id} not found")
            
            base_amount = int(product["price"]["value"] * 100 * item.quantity)
            discount = 0
            subtotal = base_amount - discount
            tax = int(subtotal * 0.1)
            total = subtotal + tax
            total_amount += total
            
            line_items.append({
                "id": f"line_item_{item.id}_{item.quantity}",
                "item": {"id": item.id, "quantity": item.quantity},
                "base_amount": base_amount,
                "discount": discount,
                "subtotal": subtotal,
                "tax": tax,
                "total": total
            })
        
        session["line_items"] = line_items
    
    if req.fulfillment_address is not None:
        session["fulfillment_address"] = req.fulfillment_address
        session["status"] = "ready_for_payment"
    
    if req.fulfillment_option_id is not None:
        session["fulfillment_option_id"] = req.fulfillment_option_id
    
    # Recalculate totals
    shipping_cost = 550 if session["fulfillment_address"] else 0
    if req.fulfillment_option_id == "express_shipping":
        shipping_cost = 1650
    
    item_total = sum(item["total"] for item in session["line_items"])
    tax_total = int(item_total * 0.1)
    grand_total = item_total + tax_total + shipping_cost
    
    session["totals"] = [
        {"type": "items_base_amount", "display_text": "Item(s) total", "amount": item_total},
        {"type": "subtotal", "display_text": "Subtotal", "amount": item_total},
        {"type": "tax", "display_text": "Tax", "amount": tax_total},
        {"type": "fulfillment", "display_text": "Shipping", "amount": shipping_cost},
        {"type": "total", "display_text": "Total", "amount": grand_total}
    ]
    
    # Store updated session
    DB["checkout_sessions"][session_id] = session
    
    audit("checkout_session_updated", {"session_id": session_id, "request": req.dict()})
    
    return session


@app.get("/checkout_sessions/{session_id}", response_model=CheckoutSession)
async def get_checkout_session(
    session_id: str,
    request: Request,
    authorization: Optional[str] = Header(None),
    x_acp_sandbox: Optional[str] = Header(None),
    sandbox: Optional[int] = Query(None)
):
    """Retrieve a checkout session"""
    require_api_key(authorization)
    sb = is_sandbox(request, x_acp_sandbox, sandbox)
    
    session = DB["checkout_sessions"].get(session_id)
    if not session:
        raise HTTPException(404, detail="Checkout session not found")
    
    return session


@app.post("/checkout_sessions/{session_id}/complete", response_model=CheckoutSessionWithOrder)
async def complete_checkout_session(
    session_id: str,
    req: CheckoutSessionCompleteRequest,
    request: Request,
    authorization: Optional[str] = Header(None),
    x_acp_sandbox: Optional[str] = Header(None),
    sandbox: Optional[int] = Query(None)
):
    """Complete a checkout session and create an order"""
    require_api_key(authorization)
    sb = is_sandbox(request, x_acp_sandbox, sandbox)
    
    # Find checkout session
    session = DB["checkout_sessions"].get(session_id)
    if not session:
        raise HTTPException(404, detail="Checkout session not found")
    
    if session["status"] in ["completed", "canceled"]:
        raise HTTPException(400, detail="Checkout session already completed or canceled")
    
    # Create order
    order_id = f"order_{int(time.time() * 1000)}"
    order = {
        "id": order_id,
        "checkout_session_id": session_id,
        "permalink_url": f"https://example.com/orders/{order_id}"
    }
    
    # Update session status
    session["status"] = "completed"
    session["buyer"] = req.buyer or session.get("buyer")
    
    # Store order and update session
    DB["orders"][order_id] = {
        **order,
        "status": "created",
        "items": session["line_items"],
        "buyer": session["buyer"],
        "fulfillment_address": session["fulfillment_address"],
        "payment_data": req.payment_data.dict(),
        "sandbox": sb,
        "created_at": datetime.now().isoformat()
    }
    
    DB["checkout_sessions"][session_id] = session
    
    audit("checkout_session_completed", {"session_id": session_id, "order_id": order_id})
    
    return {**session, "order": order}


@app.post("/checkout_sessions/{session_id}/cancel", response_model=CheckoutSession)
async def cancel_checkout_session(
    session_id: str,
    request: Request,
    authorization: Optional[str] = Header(None),
    x_acp_sandbox: Optional[str] = Header(None),
    sandbox: Optional[int] = Query(None)
):
    """Cancel a checkout session"""
    require_api_key(authorization)
    sb = is_sandbox(request, x_acp_sandbox, sandbox)
    
    session = DB["checkout_sessions"].get(session_id)
    if not session:
        raise HTTPException(404, detail="Checkout session not found")
    
    if session["status"] in ["completed", "canceled"]:
        raise HTTPException(405, detail="Checkout session cannot be canceled")
    
    # Update session status
    session["status"] = "canceled"
    session["messages"] = [{
        "type": "info",
        "param": "$",
        "content_type": "plain",
        "content": "Checkout session has been canceled."
    }]
    
    DB["checkout_sessions"][session_id] = session
    
    audit("checkout_session_canceled", {"session_id": session_id})
    
    return session


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
