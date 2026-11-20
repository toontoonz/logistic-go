# Lane

> **สรุปภาษาไทย:** Lane คือระบบจัดการพัสดุตัวเล็ก ๆ สำหรับฝึกเขียน Go
> มี REST API สำหรับออเดอร์ คนขับ และพิกัด GPS
> ประวัติตำแหน่งเก็บใน PostgreSQL/PostGIS, ตำแหน่งล่าสุดของแต่ละคนเก็บใน Redis
> แล้วส่งต่อไปที่แผนที่ของ Next.js แบบเรียลไทม์ผ่าน WebSocket
> รันทั้งระบบด้วย `docker compose up --build` แล้วเปิดหน้าแผนที่ (`/`), หน้าออเดอร์ (`/orders`),
> ตัวจำลองคนขับ (`/simulate`) และหน้าลูกค้า (`/track/<orderId>`)
> ระบบนี้ยังไม่มีการล็อกอิน — ตั้งใจไว้เป็นขั้นถัดไป และไม่ควรเปิดสู่อินเทอร์เน็ตสาธารณะ

Lane is a small parcel desk: create orders, assign a driver, and watch that driver move on a map. The Go API is the part meant for reading and changing. The Next.js app is a dispatch screen plus a public tracking page.

There is no authentication. Anyone who can reach the API can create orders and publish GPS. Run it on your machine only.

## Run it

Docker Compose starts PostGIS, Redis, the API, and the web app.

```bash
docker compose up --build
```

Then open:

| Page | URL | What it is |
| --- | --- | --- |
| Live map | http://localhost:3000 | Drivers and their latest fixes |
| Orders | http://localhost:3000/orders | Create a parcel and assign a driver |
| Simulator | http://localhost:3000/simulate | Pretends to be a phone sending GPS |
| Tracking | http://localhost:3000/track/`<orderId>` | What the customer sees |

A short demo:

1. Open the simulator and press **Start demo route**. It creates a driver named Mali, an order from Siam to Thong Lo, and asks the API to publish coordinates. The loop runs in the Go process, so it keeps going after you leave the page.
2. Open the live map (same tab is fine). Mali's marker should move.
3. Use the tracking link on the simulator page. The same driver shows up next to pickup and dropoff, and the status reads "In transit".
4. Press **Stop** on the simulator when you are done. Restarting the API also stops the loop.

Stop and delete the database volume with:

```bash
docker compose down -v
```

Ports used on the host: `3000` (web), `8080` (API), `5432` (Postgres), `6379` (Redis).

`NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_WS_URL` are baked in when the frontend image is built. The defaults point at `localhost`, which is correct when the browser runs on the same machine as Compose.

## Tests and a local frontend build

```bash
cd backend && go test ./...
cd frontend && npm install && npm run build
```

Or `make test`.

To run the API on the host instead of in a container, start only `db` and `redis` from Compose, then:

```bash
cd backend && go run ./cmd/server
cd frontend && npm run dev
```

The API reads `DATABASE_URL`, `REDIS_URL`, `HTTP_ADDR`, `CORS_ORIGINS`, and `MIGRATIONS_DIR`. Defaults match the published Compose ports. See `.env.example`.

## Layout

```
backend/          Go API
  cmd/server      process entrypoint
  internal/       config, HTTP, services, SQL, Redis, websocket hub
  migrations/     SQL applied on startup
frontend/         Next.js App Router UI
docker-compose.yml
```

Requests flow like this:

1. The browser calls the Go API for orders and drivers.
2. A GPS post is inserted into `location_history` (PostGIS geography), the latest fix is written to Redis, and the hub broadcasts it.
3. Open maps subscribe to `ws://localhost:8080/ws/positions`. On connect they receive a snapshot, then one message per ping.

## HTTP API

All JSON bodies are wrapped as `{"data": ...}` or `{"error": "..."}`.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health` | Postgres and Redis are reachable |
| `POST` | `/api/drivers` | Create a driver |
| `GET` | `/api/drivers` | List drivers, with latest fix when one exists |
| `GET` | `/api/drivers/{id}` | One driver |
| `POST` | `/api/drivers/{id}/location` | Body `{"lat":13.7,"lng":100.5}` |
| `GET` | `/api/drivers/{id}/history?limit=100` | Trail, oldest first |
| `POST` | `/api/orders` | Create a pending order |
| `GET` | `/api/orders` | List orders |
| `GET` | `/api/orders/{id}` | Order plus the driver's latest fix |
| `POST` | `/api/orders/{id}/assign` | Body `{"driverId":"..."}` |
| `PATCH` | `/api/orders/{id}/status` | One step: `picked_up`, then `in_transit`, then `delivered` |
| `GET` | `/ws/positions` | WebSocket of live positions |
| `POST` | `/api/demo/start` | Body `{"driverId":"..."}`. Server posts a Bangkok route until stop |
| `POST` | `/api/demo/stop` | Stop that loop |
| `GET` | `/api/demo` | Whether the loop is running |

Assignment is only allowed while the order is `pending` or `assigned`. Status changes cannot skip a step. `assigned` is set by the assign endpoint, not by the status endpoint, so an order cannot become assigned without a driver.

Example:

```bash
curl -s http://localhost:8080/health
curl -s -X POST http://localhost:8080/api/drivers \
  -H 'Content-Type: application/json' \
  -d '{"name":"Mali","phone":"0812345678","vehicle":"Scooter"}'
```

## Notes for reading the Go

The comments in the code stay next to the idea they describe. Three places are worth opening first:

- `cmd/server/main.go` — `signal.NotifyContext` cancels a `context.Context` on Ctrl+C or Docker stop. That context is passed into startup so a stuck database dial does not ignore shutdown. Handlers then use `context.WithTimeout` so one query cannot hold a connection forever.
- `internal/ws/hub.go` — `go hub.Run(ctx)` is a goroutine that owns the set of browsers. HTTP handlers do not touch that map. They send on channels (`register`, `unregister`, `broadcast`) and `Run` does the writes. A channel send blocks until something receives; the `done` channel in `Register` is how the caller knows the client was actually added.
- `internal/service/order.go` — `OrderStore` is an interface. Postgres implements it, and the tests implement it with a map. Go does not require an `implements` keyword; matching methods are enough. That is why `go test` needs no database.

`ST_MakePoint` in the location repository takes longitude, then latitude.

## Next steps

- Authentication and a real driver session, so GPS posts are not anonymous.
- Authorization on `/track/{id}` (a tracking token, not the raw id).
- A radius query with PostGIS (`ST_DWithin`) now that the trail is stored as `geography`.
- A phone client to replace the simulator.
