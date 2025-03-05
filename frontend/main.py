from fastapi import FastAPI, Request, Form
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
import httpx

app = FastAPI()
templates = Jinja2Templates(directory="templates")

API_URL = "http://app:8080/api"

@app.get("/", response_class=HTMLResponse)
async def home(request: Request):
    return templates.TemplateResponse("index.html", {"request": request})

# Пользовательские действия
@app.post("/user/login")
async def user_login(telegram_id: str = Form(...)):
    async with httpx.AsyncClient() as client:
        response = await client.post(f"{API_URL}/consumers/token/", json={"telegram_id": telegram_id})
        return response.json()  # Возвращаем данные JSON

@app.post("/user/ticket")
async def create_ticket(token: str = Form(...), subject: str = Form(...), description: str = Form(...), source: str = Form(...)):
    headers = {"Authorization": f"Bearer {token}"}
    data = {"subject": subject, "description": description, "source": source}
    async with httpx.AsyncClient() as client:
        response = await client.post(f"{API_URL}/tickets/create", headers=headers, json=data)
        return response.json()

@app.post("/user/message")
async def send_user_message(token: str = Form(...), ticket_id: str = Form(...), content: str = Form(...)):
    headers = {"Authorization": f"Bearer {token}"}
    data = {"sender": "user", "recipient": "operator", "content": content}
    async with httpx.AsyncClient() as client:
        response = await client.post(f"{API_URL}/tickets/{ticket_id}/messages/", headers=headers, json=data)
        return response.json()

@app.get("/user/tickets")
async def get_user_tickets(token: str):
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{API_URL}/tickets/", headers=headers)
        return response.json()

# Операторские действия
@app.post("/operator/login")
async def operator_login(username: str = Form(...), password: str = Form(...)):
    async with httpx.AsyncClient() as client:
        response = await client.post(f"{API_URL}/token/", json={"username": username, "password": password})
        return response.json()

@app.get("/operator/tickets")
async def get_all_tickets(token: str):
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{API_URL}/tickets/", headers=headers)
        return response.json()

@app.get("/operator/ticket/{ticket_id}/messages")
async def get_ticket_messages(ticket_id: str, token: str):
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{API_URL}/tickets/{ticket_id}/messages/", headers=headers)
        return response.json()

@app.post("/operator/message")
async def send_operator_message(token: str = Form(...), ticket_id: str = Form(...), content: str = Form(...)):
    headers = {"Authorization": f"Bearer {token}"}
    data = {"sender": "operator", "recipient": "user", "content": content}
    async with httpx.AsyncClient() as client:
        response = await client.post(f"{API_URL}/tickets/{ticket_id}/messages/", headers=headers, json=data)
        return response.json()

@app.post("/operator/logout")
async def operator_logout(token: str = Form(...)):
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.post(f"{API_URL}/logout/", headers=headers)
        return response.json()