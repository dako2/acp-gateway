import express from 'express'
import crypto from 'crypto'
import morgan from 'morgan'
import cors from 'cors'
import bodyParser from 'body-parser'
import multer from 'multer'
import { z } from 'zod'

// ------------------ Config ------------------
const API_KEYS = new Set((process.env.ACP_API_KEYS || 'test_key_123,another_key').split(',').map(k => k.trim()))
const WEBHOOK_SECRET = process.env.ACP_WEBHOOK_SECRET || 'whsec_123'

// ------------------ App ------------------
const app = express()
const upload = multer({ storage: multer.memoryStorage() })

// ------------------ Middleware ------------------
app.use(morgan('combined'))
app.use(cors())
app.use(bodyParser.json({ limit: '10mb' }))
app.use(bodyParser.urlencoded({ extended: true }))

// ------------------ Storage (in-mem demo) ------------------
interface DB {
  products: Product[]
  carts: Record<string, Cart>
  orders: Record<string, Order>
  merchants: Record<string, Merchant>
  audit: AuditLog[]
}

const db: DB = {
  products: [
    {
      id: "sku_123",
      title: "Work Boots",
      description: "Durable work boots for construction and outdoor work",
      variants: [{ id: "var_9", size: "9", price: { value: 119.99, currency: "USD" } }],
      price: { value: 119.99, currency: "USD" },
      available: true,
      images: ["https://example.com/boots.jpg"],
      tags: ["work", "boots", "durable"]
    },
    {
      id: "sku_456",
      title: "Safety Helmet",
      description: "ANSI approved safety helmet",
      variants: [{ id: "var_l", size: "Large", price: { value: 45.99, currency: "USD" } }],
      price: { value: 45.99, currency: "USD" },
      available: true,
      images: ["https://example.com/helmet.jpg"],
      tags: ["safety", "helmet", "ansi"]
    }
  ],
  carts: {},
  orders: {},
  merchants: {},
  audit: []
}

// ------------------ Types ------------------
interface Money {
  value: number
  currency: string
}

interface ProductVariant {
  id: string
  size?: string
  price: Money
}

interface Product {
  id: string
  title: string
  description?: string
  variants: ProductVariant[]
  price: Money
  available: boolean
  images?: string[]
  tags?: string[]
}

interface CartItem {
  product_id: string
  variant_id?: string
  quantity: number
  price?: Money
}

interface Cart {
  id: string
  items: CartItem[]
}

interface Order {
  id: string
  cart_id: string
  email?: string
  items: CartItem[]
  sandbox: boolean
  status: string
  shipping_address?: Record<string, any>
  payment?: Record<string, any>
}

interface Merchant {
  id: string
  name: string
  domain: string
  api_base?: string
  metadata?: Record<string, any>
}

interface AuditLog {
  ts: number
  event: string
  data: Record<string, any>
}

// ------------------ Schemas ------------------
const MoneySchema = z.object({
  value: z.number(),
  currency: z.string().default("USD")
})

const CartItemSchema = z.object({
  product_id: z.string(),
  variant_id: z.string().optional(),
  quantity: z.number().default(1),
  price: MoneySchema.optional()
})

const IntentPayloadSchema = z.object({
  action: z.string(),
  items: z.array(CartItemSchema).default([]),
  notes: z.string().optional(),
  context: z.record(z.any()).optional()
})

const IntentSchema = z.object({
  type: z.literal("acp.intent"),
  actor: z.literal("llm"),
  payload: IntentPayloadSchema
})

const CheckoutReqSchema = z.object({
  cart_id: z.string(),
  email: z.string().optional(),
  shipping_address: z.record(z.any()).optional(),
  payment: z.record(z.any()).optional()
})

const MerchantRegisterSchema = z.object({
  name: z.string(),
  domain: z.string(),
  api_base: z.string().optional(),
  metadata: z.record(z.any()).optional()
})

// ------------------ Helpers ------------------
function requireApiKey(auth?: string): void {
  if (!auth || !auth.startsWith("Bearer ")) {
    throw new Error("missing bearer token")
  }
  const token = auth.split(" ", 2)[1]
  if (!API_KEYS.has(token)) {
    throw new Error("invalid api key")
  }
}

function isSandbox(sandboxHeader?: string, sandboxQuery?: string): boolean {
  if (sandboxHeader && sandboxHeader.toLowerCase() === "true") return true
  if (sandboxQuery === "1") return true
  return false
}

function audit(event: string, data: Record<string, any>): void {
  db.audit.push({
    ts: Math.floor(Date.now() / 1000),
    event,
    data
  })
}

function verifyHmac(rawBody: Buffer, signature: string): void {
  try {
    const parts = Object.fromEntries(signature.split(",").map(p => p.split("=")))
    const v1 = parts.v1
    const mac = crypto.createHmac('sha256', WEBHOOK_SECRET).update(rawBody).digest('hex')
    if (!crypto.timingSafeEqual(Buffer.from(mac), Buffer.from(v1 || ""))) {
      throw new Error("invalid signature")
    }
  } catch (error) {
    throw new Error("bad signature header")
  }
}

// ------------------ Routes ------------------
app.get("/healthz", (req, res) => {
  res.json({ ok: true })
})

// API Documentation endpoints
app.get("/docs", (req, res) => {
  res.setHeader('Content-Type', 'text/html')
  const docsHTML = `<!DOCTYPE html>
<html>
<head>
    <title>ACP Gateway API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .method { font-weight: bold; color: #007acc; }
        .path { font-family: monospace; background: #e8e8e8; padding: 2px 5px; }
    </style>
</head>
<body>
    <h1>ACP Gateway API Documentation</h1>
    <p>Agentic Commerce Protocol (ACP) Gateway - Express Implementation</p>
    
    <h2>Endpoints</h2>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/healthz</span>
        <p>Health check endpoint</p>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/acp/v1/products</span>
        <p>List products with optional search and pagination</p>
        <p><strong>Headers:</strong> Authorization: Bearer &lt;API_KEY&gt;</p>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/acp/v1/intent</span>
        <p>Process intent requests from LLM agents</p>
        <p><strong>Headers:</strong> Authorization: Bearer &lt;API_KEY&gt;</p>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/acp/v1/checkout</span>
        <p>Process checkout requests</p>
        <p><strong>Headers:</strong> Authorization: Bearer &lt;API_KEY&gt;</p>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/merchant/register</span>
        <p>Register new merchants</p>
        <p><strong>Headers:</strong> Authorization: Bearer &lt;API_KEY&gt;</p>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/merchant/feed</span>
        <p>Upload product feeds</p>
        <p><strong>Headers:</strong> Authorization: Bearer &lt;API_KEY&gt;</p>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/webhooks/acp</span>
        <p>Receive webhook notifications</p>
        <p><strong>Headers:</strong> X-Acp-Signature: t=&lt;ts&gt;,v1=&lt;hex&gt;</p>
    </div>
    
    <h2>Authentication</h2>
    <p>Most endpoints require API key authentication via the Authorization header:</p>
    <code>Authorization: Bearer your_api_key</code>
    
    <h2>More Information</h2>
    <p>For detailed API documentation, see <a href="/openapi.json">OpenAPI Specification</a></p>
    <p>For implementation details, see the <a href="https://github.com/agentic-commerce-protocol">ACP Protocol</a> repository.</p>
</body>
</html>`
  res.send(docsHTML)
})

app.get("/openapi.json", (req, res) => {
  res.json({
    openapi: "3.0.0",
    info: {
      title: "ACP Gateway",
      version: "0.1.0",
      description: "Agentic Commerce Protocol (ACP) Gateway - Express Implementation"
    },
    servers: [
      { url: "http://localhost:8081", description: "Express implementation" }
    ],
    paths: {
      "/healthz": {
        get: {
          summary: "Health check",
          description: "Returns the health status of the gateway",
          responses: {
            "200": {
              description: "OK",
              content: {
                "application/json": {
                  schema: {
                    type: "object",
                    properties: {
                      ok: { type: "boolean" }
                    }
                  }
                }
              }
            }
          }
        }
      },
      "/acp/v1/products": {
        get: {
          summary: "List products",
          description: "Retrieve a list of available products",
          security: [{ ApiKeyAuth: [] }],
          parameters: [
            { name: "q", in: "query", schema: { type: "string" } },
            { name: "limit", in: "query", schema: { type: "integer", default: 20 } },
            { name: "offset", in: "query", schema: { type: "integer", default: 0 } }
          ]
        }
      }
    },
    components: {
      securitySchemes: {
        ApiKeyAuth: {
          type: "http",
          scheme: "bearer"
        }
      }
    }
  })
})

app.get("/acp/v1/products", (req, res) => {
  try {
    requireApiKey(req.headers.authorization)
    
    const { q, limit = "20", offset = "0" } = req.query
    let items = db.products
    
    if (q && typeof q === "string") {
      const query = q.toLowerCase()
      items = items.filter(p => 
        p.title.toLowerCase().includes(query) || 
        p.id.toLowerCase().includes(query) ||
        (p.description && p.description.toLowerCase().includes(query))
      )
    }
    
    const start = parseInt(offset as string)
    const end = start + parseInt(limit as string)
    
    res.json({
      ok: true,
      items: items.slice(start, end),
      total: items.length
    })
  } catch (error) {
    res.status(401).json({ error: error instanceof Error ? error.message : "unauthorized" })
  }
})

app.post("/acp/v1/intent", (req, res) => {
  try {
    requireApiKey(req.headers.authorization)
    
    const sandbox = isSandbox(req.headers['x-acp-sandbox'] as string, req.query.sandbox as string)
    const intent = IntentSchema.parse(req.body)
    
    // init cart
    const cartId = intent.payload.context?.session_id || `cart_${Date.now()}`
    let cart = db.carts[cartId]
    if (!cart) {
      cart = { id: cartId, items: [] }
      db.carts[cartId] = cart
    }
    
    if (intent.payload.action === "add_to_cart") {
      for (const item of intent.payload.items) {
        const product = db.products.find(p => p.id === item.product_id)
        if (!product) {
          throw new Error(`product ${item.product_id} not found`)
        }
        const price = item.price || product.price
        cart.items.push({
          product_id: item.product_id,
          variant_id: item.variant_id,
          quantity: item.quantity,
          price
        })
      }
    } else if (intent.payload.action === "select_variant") {
      for (const item of intent.payload.items) {
        const cartItem = cart.items.find(ci => ci.product_id === item.product_id)
        if (cartItem) {
          cartItem.variant_id = item.variant_id
        }
      }
    } else if (intent.payload.action === "set_quantity") {
      for (const item of intent.payload.items) {
        const cartItem = cart.items.find(ci => ci.product_id === item.product_id)
        if (cartItem) {
          cartItem.quantity = item.quantity
        }
      }
    } else if (intent.payload.action === "checkout") {
      // defer to /acp/v1/checkout
    } else {
      throw new Error("unknown action")
    }
    
    const subtotal = cart.items.reduce((sum, item) => 
      sum + (item.price?.value || 0) * item.quantity, 0
    )
    
    const result = {
      type: "acp.intent.result",
      ok: true,
      sandbox,
      cart: {
        ...cart,
        subtotal: { value: Math.round(subtotal * 100) / 100, currency: "USD" }
      },
      next: { action: "checkout", endpoint: "/acp/v1/checkout" }
    }
    
    audit("intent", { sandbox, intent: req.body, result })
    res.json(result)
    
  } catch (error) {
    const status = error instanceof Error && error.message.includes("not found") ? 404 : 400
    res.status(status).json({ error: error instanceof Error ? error.message : "bad request" })
  }
})

app.post("/acp/v1/checkout", (req, res) => {
  try {
    requireApiKey(req.headers.authorization)
    
    const sandbox = isSandbox(req.headers['x-acp-sandbox'] as string, req.query.sandbox as string)
    const checkoutReq = CheckoutReqSchema.parse(req.body)
    
    const cart = db.carts[checkoutReq.cart_id]
    if (!cart) {
      throw new Error("cart not found")
    }
    
    const orderId = `ord_${Date.now()}`
    db.orders[orderId] = {
      id: orderId,
      cart_id: checkoutReq.cart_id,
      email: checkoutReq.email,
      items: cart.items,
      sandbox,
      status: "created",
      shipping_address: checkoutReq.shipping_address,
      payment: checkoutReq.payment
    }
    
    const response = {
      ok: true,
      sandbox,
      order_id: orderId,
      payment_intent_status: sandbox ? "requires_action" : "succeeded",
      redirect_url: `https://merchant.example/checkout/${orderId}`
    }
    
    audit("checkout", { sandbox, req: checkoutReq, resp: response })
    res.json(response)
    
  } catch (error) {
    const status = error instanceof Error && error.message.includes("not found") ? 404 : 400
    res.status(status).json({ error: error instanceof Error ? error.message : "bad request" })
  }
})

app.post("/merchant/register", (req, res) => {
  try {
    requireApiKey(req.headers.authorization)
    
    const merchantData = MerchantRegisterSchema.parse(req.body)
    const merchantId = merchantData.domain
    
    db.merchants[merchantId] = {
      id: merchantId,
      ...merchantData
    }
    
    audit("merchant.register", merchantData)
    res.json({ ok: true, merchant_id: merchantId })
    
  } catch (error) {
    res.status(400).json({ error: error instanceof Error ? error.message : "bad request" })
  }
})

app.post("/merchant/feed", upload.single('file'), (req, res) => {
  try {
    requireApiKey(req.headers.authorization)
    
    const { url } = req.query
    
    // Demo: wipe + seed with one product or parse uploaded file
    db.products = db.products.slice(0, 1)
    
    if (req.file) {
      const content = req.file.buffer.toString('utf-8').slice(0, 5000) // truncate demo
      audit("merchant.feed.upload", { size: content.length })
    } else if (url) {
      audit("merchant.feed.url", { url })
    }
    
    res.json({ ok: true, ingested: db.products.length })
    
  } catch (error) {
    res.status(400).json({ error: error instanceof Error ? error.message : "bad request" })
  }
})

app.post("/webhooks/acp", (req, res) => {
  try {
    const signature = req.headers['x-acp-signature'] as string
    if (!signature) {
      throw new Error("missing signature")
    }
    
    verifyHmac(req.body, signature)
    const payload = req.body
    
    audit("webhook", payload)
    res.json({ ok: true })
    
  } catch (error) {
    const status = error instanceof Error && error.message.includes("missing") ? 401 : 400
    res.status(status).json({ error: error instanceof Error ? error.message : "bad request" })
  }
})

// ------------------ Start Server ------------------
const PORT = process.env.PORT || 8080
app.listen(PORT, () => {
  console.log(`ACP Gateway Express server running on port ${PORT}`)
})
