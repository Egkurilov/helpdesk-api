from fastapi import FastAPI, Request, Form, Header, HTTPException
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
import httpx
from pydantic import BaseModel
import logging

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI()
templates = Jinja2Templates(directory="templates")
API_URL = "http://app:8080/api"

def get_token(authorization: str = Header(...)):
    if not authorization.startswith("Bearer "):
        raise HTTPException(status_code=400, detail="Invalid authorization header")
    return authorization[7:]

@app.post("/operator/login")
async def operator_login(username: str = Form(...), password: str = Form(...)):
    logger.info(f"Sending login request to {API_URL}/token/ with username={username}")
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{API_URL}/token/",
            json={"username": username, "password": password}
        )
        logger.info(f"Go API response: {response.status_code} - {response.text}")
        return response.json()

@app.get("/operator/tickets")
async def get_all_tickets(authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    logger.info(f"Sending request to {API_URL}/tickets/ with token: {token[:10]}...")
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{API_URL}/tickets/", headers=headers)
        logger.info(f"Go API response: {response.status_code} - {response.text}")
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
        response = await client.get(f"{API_URL}/operator/whitelist/", headers=headers)
        return response.json()

class WhitelistEdit(BaseModel):
    permission: str

@app.post("/operator/whitelist/{telegram_id}/edit")
async def edit_whitelist(
        telegram_id: str,
        edit: WhitelistEdit,
        authorization: str = Header(...)
):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    data = {"permission": edit.permission}
    logger.info(f"Sending POST request to {API_URL}/operator/whitelist/{telegram_id}/edit with data: {data}")
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{API_URL}/operator/whitelist/{telegram_id}/edit",
            headers=headers,
            json=data
        )
        logger.info(f"Go API response: {response.status_code} - {response.text}")
        if response.status_code != 200:
            raise HTTPException(status_code=response.status_code, detail=response.text)
        return response.json()

@app.get("/operator/whitelist/all")
async def get_whitelist_all(authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{API_URL}/operator/whitelist/all", headers=headers)
        return response.json()

@app.get("/operator/settings/")
async def get_settings(authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    async with httpx.AsyncClient() as client:
        try:
            response = await client.get(f"{API_URL}/operator/settings/", headers=headers)
            response.raise_for_status()
            logger.info(f"GET /operator/settings/ response: {response.status_code} - {response.text}")
            return response.json()
        except httpx.RequestError as exc:
            logger.error(f"Cannot connect to Go API: {str(exc)}")
            raise HTTPException(status_code=503, detail=f"Cannot connect to Go API: {str(exc)}")

@app.post("/operator/settings/")
async def add_endpoint(request: Request, authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    data = await request.json()
    logger.info(f"Received POST /operator/settings/ with data: {data}")
    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(f"{API_URL}/operator/settings/", headers=headers, json=data)
            response.raise_for_status()
            logger.info(f"POST /operator/settings/ response: {response.status_code} - {response.text}")
            return response.json()
        except httpx.RequestError as exc:
            logger.error(f"Cannot connect to Go API: {str(exc)}")
            raise HTTPException(status_code=503, detail=f"Cannot connect to Go API: {str(exc)}")
        except httpx.HTTPStatusError as exc:
            logger.error(f"Go API error: {exc.response.status_code} - {exc.response.text}")
            raise HTTPException(status_code=exc.response.status_code, detail=exc.response.text)

@app.put("/operator/settings/{id}")
async def update_endpoint(id: str, request: Request, authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    data = await request.json()
    logger.info(f"Received PUT /operator/settings/{id} with data: {data}")
    async with httpx.AsyncClient() as client:
        try:
            response = await client.put(f"{API_URL}/operator/settings/{id}", headers=headers, json=data)
            response.raise_for_status()
            logger.info(f"PUT /operator/settings/{id} response: {response.status_code} - {response.text}")
            return response.json()
        except httpx.RequestError as exc:
            logger.error(f"Cannot connect to Go API: {str(exc)}")
            raise HTTPException(status_code=503, detail=f"Cannot connect to Go API: {str(exc)}")
        except httpx.HTTPStatusError as exc:
            logger.error(f"Go API error: {exc.response.status_code} - {exc.response.text}")
            raise HTTPException(status_code=exc.response.status_code, detail=exc.response.text)

@app.delete("/operator/settings/{id}")
async def delete_endpoint(id: str, authorization: str = Header(...)):
    token = get_token(authorization)
    headers = {"Authorization": f"Bearer {token}"}
    logger.info(f"Received DELETE /operator/settings/{id}")
    async with httpx.AsyncClient() as client:
        try:
            response = await client.delete(f"{API_URL}/operator/settings/{id}", headers=headers)
            response.raise_for_status()
            logger.info(f"DELETE /operator/settings/{id} response: {response.status_code}")
            return {"message": "Endpoint deleted"}
        except httpx.RequestError as exc:
            logger.error(f"Cannot connect to Go API: {str(exc)}")
            raise HTTPException(status_code=503, detail=f"Cannot connect to Go API: {str(exc)}")
        except httpx.HTTPStatusError as exc:
            logger.error(f"Go API error: {exc.response.status_code} - {exc.response.text}")
            raise HTTPException(status_code=exc.response.status_code, detail=exc.response.text)

@app.get("/{path:path}", response_class=HTMLResponse)
async def home(request: Request, path: str):
    return templates.TemplateResponse("index.html", {"request": request})