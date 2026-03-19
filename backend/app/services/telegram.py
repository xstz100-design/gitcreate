"""Telegram Bot 通知服务"""
import httpx
import logging
from typing import Optional

logger = logging.getLogger(__name__)


async def send_telegram_message(bot_token: str, chat_id: str, message: str) -> bool:
    """发送 Telegram 消息"""
    if not bot_token or not chat_id:
        return False
    
    url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
    payload = {
        "chat_id": chat_id,
        "text": message,
        "parse_mode": "HTML",
    }
    
    try:
        async with httpx.AsyncClient(timeout=10) as client:
            resp = await client.post(url, json=payload)
            if resp.status_code == 200:
                logger.info(f"Telegram 消息发送成功: chat_id={chat_id}")
                return True
            else:
                logger.warning(f"Telegram 消息发送失败: {resp.status_code} {resp.text}")
                return False
    except Exception as e:
        logger.error(f"Telegram 消息发送异常: {e}")
        return False


def send_telegram_sync(bot_token: str, chat_id: str, message: str) -> bool:
    """同步方式发送 Telegram 消息 (用于非异步上下文)"""
    if not bot_token or not chat_id:
        return False
    
    url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
    payload = {
        "chat_id": chat_id,
        "text": message,
        "parse_mode": "HTML",
    }
    
    try:
        with httpx.Client(timeout=10) as client:
            resp = client.post(url, json=payload)
            if resp.status_code == 200:
                logger.info(f"Telegram 消息发送成功: chat_id={chat_id}")
                return True
            else:
                logger.warning(f"Telegram 消息发送失败: {resp.status_code} {resp.text}")
                return False
    except Exception as e:
        logger.error(f"Telegram 消息发送异常: {e}")
        return False


def format_new_order_message(order_no: str, merchant_name: str, total_usd: float, items: list) -> str:
    """格式化新订单通知消息"""
    item_lines = []
    for item in items:
        item_lines.append(f"  • {item['product_name']} x{item['quantity']} = ${item['subtotal_usd']:.2f}")
    
    items_text = "\n".join(item_lines)
    
    return (
        f"🛒 <b>新订单通知</b>\n"
        f"━━━━━━━━━━━━━━━\n"
        f"📋 订单号: <code>{order_no}</code>\n"
        f"👤 商户: {merchant_name}\n"
        f"💰 总金额: <b>${total_usd:.2f}</b>\n"
        f"📦 商品明细:\n{items_text}\n"
        f"━━━━━━━━━━━━━━━\n"
        f"请及时处理！"
    )


def format_low_stock_message(product_name: str, current_stock: int, warning_level: int) -> str:
    """格式化库存预警消息"""
    return (
        f"⚠️ <b>库存预警</b>\n"
        f"━━━━━━━━━━━━━━━\n"
        f"📦 商品: {product_name}\n"
        f"📉 当前库存: <b>{current_stock}</b>\n"
        f"🔔 预警值: {warning_level}\n"
        f"━━━━━━━━━━━━━━━\n"
        f"请及时补货！"
    )


def notify_admins_new_order(session, order_no: str, merchant_name: str, total_usd: float, items: list):
    """通知所有设置了 Telegram 的管理员有新订单"""
    from app.models import User, UserRole
    from sqlmodel import select
    
    admins = session.exec(
        select(User).where(
            User.role == UserRole.ADMIN,
            User.telegram_bot_token.isnot(None),  # type: ignore
            User.telegram_chat_id.isnot(None),  # type: ignore
        )
    ).all()
    
    if not admins:
        return
    
    message = format_new_order_message(order_no, merchant_name, total_usd, items)
    
    for admin in admins:
        if admin.telegram_bot_token and admin.telegram_chat_id:
            send_telegram_sync(admin.telegram_bot_token, admin.telegram_chat_id, message)


def notify_admins_low_stock(session, product_name: str, current_stock: int, warning_level: int):
    """通知所有设置了 Telegram 的管理员库存预警"""
    from app.models import User, UserRole
    from sqlmodel import select
    
    admins = session.exec(
        select(User).where(
            User.role == UserRole.ADMIN,
            User.telegram_bot_token.isnot(None),  # type: ignore
            User.telegram_chat_id.isnot(None),  # type: ignore
        )
    ).all()
    
    if not admins:
        return
    
    message = format_low_stock_message(product_name, current_stock, warning_level)
    
    for admin in admins:
        if admin.telegram_bot_token and admin.telegram_chat_id:
            send_telegram_sync(admin.telegram_bot_token, admin.telegram_chat_id, message)
