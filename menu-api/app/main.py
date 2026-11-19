import os
import time
from decimal import Decimal

import psycopg2
from psycopg2.extras import RealDictCursor
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

app = FastAPI(title="Dhaba Orders Menu API", version="1.0.0")

DB_HOST = os.getenv("DB_HOST", "postgres")
DB_NAME = os.getenv("DB_NAME", "dhaba")
DB_USER = os.getenv("DB_USER", "dhaba")
DB_PASSWORD = os.getenv("DB_PASSWORD", "dhaba-dev-change-this")
DB_PORT = int(os.getenv("DB_PORT", "5432"))

def get_connection():
    return psycopg2.connect(
        host=DB_HOST,
        dbname=DB_NAME,
        user=DB_USER,
        password=DB_PASSWORD,
        port=DB_PORT,
        connect_timeout=3,
    )

class MenuItemCreate(BaseModel):
    name: str = Field(min_length=2, max_length=120)
    description: str = ""
    price: Decimal = Field(gt=0, max_digits=10, decimal_places=2)

@app.get("/health")
def health():
    try:
        conn = get_connection()
        conn.close()
        return {"status": "healthy", "database": "connected"}
    except Exception as exc:
        raise HTTPException(status_code=503, detail=f"Database unavailable: {str(exc)}")

@app.get("/menu")
def list_menu():
    try:
        with get_connection() as conn:
            with conn.cursor(cursor_factory=RealDictCursor) as cur:
                cur.execute("""
                    SELECT id, name, description, price, available
                    FROM menu_items
                    WHERE available = TRUE
                    ORDER BY id
                """)
                return cur.fetchall()
    except Exception as exc:
        raise HTTPException(status_code=503, detail=str(exc))

@app.get("/menu/{item_id}")
def get_menu_item(item_id: int):
    with get_connection() as conn:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute("""
                SELECT id, name, description, price, available
                FROM menu_items WHERE id = %s
            """, (item_id,))
            item = cur.fetchone()
            if not item:
                raise HTTPException(status_code=404, detail="Menu item not found")
            return item

@app.post("/menu", status_code=201)
def create_menu_item(item: MenuItemCreate):
    try:
        with get_connection() as conn:
            with conn.cursor(cursor_factory=RealDictCursor) as cur:
                cur.execute("""
                    INSERT INTO menu_items (name, description, price)
                    VALUES (%s, %s, %s)
                    RETURNING id, name, description, price, available
                """, (item.name, item.description, item.price))
                return cur.fetchone()
    except psycopg2.errors.UniqueViolation:
        raise HTTPException(status_code=409, detail="Menu item already exists")
    except Exception as exc:
        raise HTTPException(status_code=503, detail=str(exc))
