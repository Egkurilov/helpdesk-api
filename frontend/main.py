from fastapi import FastAPI, Request, Form, Header, HTTPException
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
import httpx

app = FastAPI()
templates = Jinja2Templates(directory="templates")

# API_URL должен включать префикс API, как ожидает Go-сервис
API_URL = "http://app:8080/api"  # При необходимости отредактируйте это значение

def get_token(authorization: str = Header(...)):
    if not authorization.startswith("Bearer "):
        raise HTTPException(status_code=400, detail="Invalid authorization header")
    return authorization[7:]

@app.post("/operator/login")
async def operator_login(username: str = Form(...), password: str = Form(...)):
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{API_URL}/token/",
            json={"username": username, "password": password}
        )
        return response.json()

@app.get("/operator/tickets")
async def get_all_tickets(authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{API_URL}/tickets/", headers=headers)
        return response.json()

@app.get("/operator/ticket/{ticket_id}/messages")
async def get_ticket_messages(ticket_id: str, authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.get(
            f"{API_URL}/tickets/{ticket_id}/messages/",
            headers=headers
        )
        return response.json()

@app.post("/operator/message")
async def send_operator_message(
    token: str = Form(...),
    ticket_id: str = Form(...),
    content: str = Form(...)
):
    headers = {"Authorization": f"Bearer {token}"}
    data = {"sender": "operator", "recipient": "user", "content": content}
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{API_URL}/tickets/{ticket_id}/messages/",
            headers=headers,
            json=data
        )
        return response.json()

@app.post("/operator/logout")
async def operator_logout(token: str = Form(...)):
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.post(f"{API_URL}/logout/", headers=headers)
        return response.json()

@app.post("/operator/ticket/{ticket_id}/close")
async def close_ticket(ticket_id: str, token: str = Form(...)):
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{API_URL}/tickets/{ticket_id}/close/",
            headers=headers
        )
        return response.json()

@app.get("/operator/whitelist")
async def get_whitelist_pending(authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        # Запрос к эндпоинту для Pending-записей
        response = await client.get(f"{API_URL}/operator/whitelist/", headers=headers)
        return response.json()

@app.post("/operator/whitelist/{telegram_id}/edit")
async def edit_whitelist(
    telegram_id: str,
    perm: str = Form(...),
    authorization: str = Header(...)
):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    # Приведение perm к булевому типу: "true" → True, иначе False
    perm_bool = True if perm.lower() == "true" else False
    data = {"perm": perm_bool}
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{API_URL}/operator/whitelist/{telegram_id}/edit",
            headers=headers,
            json=data
        )
        return response.json()

@app.get("/operator/whitelist/all")
async def get_whitelist_all(authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{API_URL}/operator/whitelist/all", headers=headers)
        return response.json()

# Catch-all маршрут для отдачи HTML-страницы (всегда последний)
@app.get("/{path:path}", response_class=HTMLResponse)
async def home(request: Request, path: str):
    return templates.TemplateResponse("index.html", {"request": request})
