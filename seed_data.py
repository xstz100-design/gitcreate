"""
Reset and re-seed all data: categories, products, announcements.
"""
import requests, warnings
warnings.filterwarnings('ignore')

BASE = "https://khmerai.cn/api"

# 1. Login
print("=== Login ===")
r = requests.post(f"{BASE}/auth/login", json={"username": "admin", "password": "admin123"}, verify=False)
if r.status_code != 200:
    print("Login failed:", r.text); exit(1)
token = r.json()["access_token"]
H = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
print("OK")

# 2. Delete announcements
print("\n=== Clear announcements ===")
ann_list = requests.get(f"{BASE}/announcements", headers=H, verify=False).json()
ann_list = ann_list if isinstance(ann_list, list) else []
for a in ann_list:
    requests.delete(f"{BASE}/announcements/{a['id']}", headers=H, verify=False)
print(f"  Deleted {len(ann_list)}")

# 3. Delete products
print("\n=== Clear products ===")
resp = requests.get(f"{BASE}/products", headers=H, verify=False).json()
prods = resp if isinstance(resp, list) else resp.get("items", resp.get("data", []))
for p in prods:
    requests.delete(f"{BASE}/products/{p['id']}", headers=H, verify=False)
print(f"  Deleted {len(prods)}")

# 4. Delete categories
print("\n=== Clear categories ===")
resp = requests.get(f"{BASE}/categories/all", headers=H, verify=False).json()
cats = resp if isinstance(resp, list) else resp.get("data", [])
for c in cats:
    requests.delete(f"{BASE}/categories/{c['id']}", headers=H, verify=False)
print(f"  Deleted {len(cats)}")

# 5. Create categories
print("\n=== Create categories ===")
CATS = [
    {"name": "Drinks / Beverages",  "sort_order": 1, "is_active": True},
    {"name": "Snacks / Zero Food",  "sort_order": 2, "is_active": True},
    {"name": "Condiments",          "sort_order": 3, "is_active": True},
    {"name": "Daily Goods",         "sort_order": 4, "is_active": True},
    {"name": "Grains & Oil",        "sort_order": 5, "is_active": True},
    {"name": "Dairy",               "sort_order": 6, "is_active": True},
    {"name": "Instant Food",        "sort_order": 7, "is_active": True},
    {"name": "Alcohol / Beer",      "sort_order": 8, "is_active": True},
]
for c in CATS:
    r = requests.post(f"{BASE}/categories", headers=H, json=c, verify=False)
    print(f"  {'OK' if r.status_code in (200,201) else 'FAIL'} {c['name']}")

# 6. Create products
print("\n=== Create products ===")
PRODUCTS = [
    # Drinks
    {"name": "Coca-Cola 330ml",     "name_kh": "កូកាកូឡា", "name_en": "Coca-Cola 330ml",
     "brand": "Coca-Cola", "country_of_origin": "Cambodia", "category": "Drinks / Beverages",
     "unit": "can", "price_usd": 0.75, "stock": 480, "stock_warning": 48,
     "pieces_per_package": 24, "pack_name": "case", "specs": "330ml x24/case",
     "is_featured": True, "is_active": True, "sort_order": 1, "description": "Classic Coca-Cola, refreshing."},
    {"name": "Sprite 330ml",        "name_kh": "ស្ព្រាយ", "name_en": "Sprite 330ml",
     "brand": "Sprite", "country_of_origin": "Cambodia", "category": "Drinks / Beverages",
     "unit": "can", "price_usd": 0.70, "stock": 480, "stock_warning": 48,
     "pieces_per_package": 24, "pack_name": "case", "specs": "330ml x24/case",
     "is_featured": True, "is_active": True, "sort_order": 2, "description": "Fresh lemon-lime soda."},
    {"name": "Mineral Water 500ml", "name_kh": "ទឹកជ្រោះ", "name_en": "Mineral Water 500ml",
     "brand": "Kulen", "country_of_origin": "Cambodia", "category": "Drinks / Beverages",
     "unit": "bottle", "price_usd": 0.30, "stock": 1200, "stock_warning": 120,
     "pieces_per_package": 24, "pack_name": "case", "specs": "500ml x24/case",
     "is_featured": False, "is_active": True, "sort_order": 3, "description": "Natural mineral water."},
    {"name": "Red Bull 250ml",      "name_kh": "រេដប៊ូល", "name_en": "Red Bull Energy Drink 250ml",
     "brand": "Red Bull", "country_of_origin": "Thailand", "category": "Drinks / Beverages",
     "unit": "can", "price_usd": 1.20, "stock": 288, "stock_warning": 24,
     "pieces_per_package": 24, "pack_name": "case", "specs": "250ml x24/case",
     "is_featured": True, "is_active": True, "sort_order": 4, "description": "Energy drink, boosts focus."},
    {"name": "Angkor Ice Tea 500ml","name_kh": "តែទឹកកក", "name_en": "Angkor Ice Tea 500ml",
     "brand": "Angkor", "country_of_origin": "Cambodia", "category": "Drinks / Beverages",
     "unit": "bottle", "price_usd": 0.55, "stock": 360, "stock_warning": 36,
     "pieces_per_package": 12, "pack_name": "case", "specs": "500ml x12/case",
     "is_featured": False, "is_active": True, "sort_order": 5, "description": "Local Cambodian iced tea."},
    # Snacks
    {"name": "Lay's Original 85g",  "name_kh": "ឡេស", "name_en": "Lay's Original Chips 85g",
     "brand": "Lay's", "country_of_origin": "Thailand", "category": "Snacks / Zero Food",
     "unit": "bag", "price_usd": 1.50, "stock": 200, "stock_warning": 20,
     "pieces_per_package": 20, "pack_name": "case", "specs": "85g x20/case",
     "is_featured": True, "is_active": True, "sort_order": 1, "description": "Classic original chips, crispy."},
    {"name": "Oreo 137g",           "name_kh": "អូរីអូ", "name_en": "Oreo Sandwich Cookies 137g",
     "brand": "Oreo", "country_of_origin": "China", "category": "Snacks / Zero Food",
     "unit": "pack", "price_usd": 1.80, "stock": 144, "stock_warning": 24,
     "pieces_per_package": 12, "pack_name": "case", "specs": "137g x12/case",
     "is_featured": False, "is_active": True, "sort_order": 2, "description": "Classic black & white cookies."},
    {"name": "Oishi Prawn Crackers 55g","name_kh": "គ្រាប់បង្គា", "name_en": "Oishi Prawn Crackers 55g",
     "brand": "Oishi", "country_of_origin": "Thailand", "category": "Snacks / Zero Food",
     "unit": "bag", "price_usd": 0.55, "stock": 400, "stock_warning": 40,
     "pieces_per_package": 40, "pack_name": "case", "specs": "55g x40/case",
     "is_featured": False, "is_active": True, "sort_order": 3, "description": "Crispy prawn crackers."},
    {"name": "Pocky Chocolate 40g", "name_kh": "ផូគ្គី", "name_en": "Pocky Chocolate Sticks 40g",
     "brand": "Glico", "country_of_origin": "Thailand", "category": "Snacks / Zero Food",
     "unit": "box", "price_usd": 0.70, "stock": 200, "stock_warning": 30,
     "pieces_per_package": 30, "pack_name": "case", "specs": "40g x30/case",
     "is_featured": False, "is_active": True, "sort_order": 4, "description": "Chocolate biscuit sticks."},
    # Condiments
    {"name": "Kimsan Soy Sauce 700ml","name_kh": "សូសសៀវ", "name_en": "Kimsan Soy Sauce 700ml",
     "brand": "Kimsan", "country_of_origin": "Cambodia", "category": "Condiments",
     "unit": "bottle", "price_usd": 1.10, "stock": 144, "stock_warning": 24,
     "pieces_per_package": 12, "pack_name": "case", "specs": "700ml x12/case",
     "is_featured": False, "is_active": True, "sort_order": 1, "description": "Naturally brewed soy sauce."},
    {"name": "Fish Sauce 700ml",    "name_kh": "ទឹកត្រី", "name_en": "Tiparos Fish Sauce 700ml",
     "brand": "Tiparos", "country_of_origin": "Thailand", "category": "Condiments",
     "unit": "bottle", "price_usd": 1.40, "stock": 144, "stock_warning": 24,
     "pieces_per_package": 12, "pack_name": "case", "specs": "700ml x12/case",
     "is_featured": False, "is_active": True, "sort_order": 2, "description": "Thai imported fish sauce."},
    {"name": "Oyster Sauce 510g",   "name_kh": "ទឹកមហូប", "name_en": "Lee Kum Kee Oyster Sauce 510g",
     "brand": "Lee Kum Kee", "country_of_origin": "China", "category": "Condiments",
     "unit": "bottle", "price_usd": 2.20, "stock": 96, "stock_warning": 12,
     "pieces_per_package": 12, "pack_name": "case", "specs": "510g x12/case",
     "is_featured": False, "is_active": True, "sort_order": 3, "description": "Premium oyster sauce, savory."},
    # Grains & Oil
    {"name": "Thai Jasmine Rice 5kg","name_kh": "អង្ករ", "name_en": "Thai Jasmine Rice 5kg",
     "brand": "Sunflower", "country_of_origin": "Thailand", "category": "Grains & Oil",
     "unit": "bag", "price_usd": 7.00, "stock": 50, "stock_warning": 10,
     "pieces_per_package": 4, "pack_name": "set", "specs": "5kg x4/set",
     "is_featured": True, "is_active": True, "sort_order": 1, "description": "Fragrant Thai jasmine rice."},
    {"name": "Peanut Oil 1L",       "name_kh": "ប្រេងសណ្ដែក", "name_en": "Golden Dragon Peanut Oil 1L",
     "brand": "Golden Dragon", "country_of_origin": "Cambodia", "category": "Grains & Oil",
     "unit": "bottle", "price_usd": 3.00, "stock": 48, "stock_warning": 6,
     "pieces_per_package": 12, "pack_name": "case", "specs": "1L x12/case",
     "is_featured": False, "is_active": True, "sort_order": 2, "description": "Pure peanut oil, rich aroma."},
    # Daily Goods
    {"name": "Tide Powder 1kg",     "name_kh": "ម្សៅលាងសំពត់", "name_en": "Tide Washing Powder 1kg",
     "brand": "Tide", "country_of_origin": "Vietnam", "category": "Daily Goods",
     "unit": "bag", "price_usd": 2.50, "stock": 100, "stock_warning": 12,
     "pieces_per_package": 10, "pack_name": "case", "specs": "1kg x10/case",
     "is_featured": False, "is_active": True, "sort_order": 1, "description": "Strong stain removal."},
    {"name": "Safeguard Soap 85g",  "name_kh": "សាប៊ូ", "name_en": "Safeguard Bar Soap 85g",
     "brand": "Safeguard", "country_of_origin": "China", "category": "Daily Goods",
     "unit": "bar", "price_usd": 0.75, "stock": 288, "stock_warning": 48,
     "pieces_per_package": 48, "pack_name": "case", "specs": "85g x48/case",
     "is_featured": False, "is_active": True, "sort_order": 2, "description": "Antibacterial soap."},
    {"name": "Pantene Shampoo 400ml","name_kh": "សាប៊ូកក", "name_en": "Pantene Shampoo 400ml",
     "brand": "Pantene", "country_of_origin": "Thailand", "category": "Daily Goods",
     "unit": "bottle", "price_usd": 2.80, "stock": 60, "stock_warning": 12,
     "pieces_per_package": 12, "pack_name": "case", "specs": "400ml x12/case",
     "is_featured": False, "is_active": True, "sort_order": 3, "description": "Moisturizing shampoo."},
    # Instant Food
    {"name": "Indomie Noodles 85g", "name_kh": "មីអាំង", "name_en": "Indomie Instant Noodles 85g",
     "brand": "Indomie", "country_of_origin": "Indonesia", "category": "Instant Food",
     "unit": "pack", "price_usd": 0.40, "stock": 800, "stock_warning": 80,
     "pieces_per_package": 40, "pack_name": "case", "specs": "85g x40/case",
     "is_featured": True, "is_active": True, "sort_order": 1, "description": "Indonesian instant noodles."},
    {"name": "Mama Noodles 55g",    "name_kh": "ម៉ាម៉ា", "name_en": "Mama Instant Noodles 55g",
     "brand": "Mama", "country_of_origin": "Thailand", "category": "Instant Food",
     "unit": "pack", "price_usd": 0.25, "stock": 1000, "stock_warning": 100,
     "pieces_per_package": 60, "pack_name": "case", "specs": "55g x60/case",
     "is_featured": True, "is_active": True, "sort_order": 2, "description": "Thai Mama noodles, affordable."},
    # Dairy
    {"name": "Bright Milk 250ml",   "name_kh": "ទឹកដោះគោ", "name_en": "Bright Pure Milk 250ml",
     "brand": "Bright Dairy", "country_of_origin": "China", "category": "Dairy",
     "unit": "box", "price_usd": 0.85, "stock": 144, "stock_warning": 24,
     "pieces_per_package": 24, "pack_name": "case", "specs": "250ml x24/case",
     "is_featured": False, "is_active": True, "sort_order": 1, "description": "Fresh pure milk, nutritious."},
    # Alcohol
    {"name": "Angkor Beer 330ml",   "name_kh": "អង្គរបៀរ", "name_en": "Angkor Beer 330ml",
     "brand": "Angkor", "country_of_origin": "Cambodia", "category": "Alcohol / Beer",
     "unit": "can", "price_usd": 0.85, "stock": 288, "stock_warning": 24,
     "pieces_per_package": 24, "pack_name": "case", "specs": "330ml x24/case",
     "is_featured": True, "is_active": True, "sort_order": 1, "description": "Cambodia's iconic beer."},
    {"name": "Tiger Beer 330ml",    "name_kh": "ហ្គ័របៀរ", "name_en": "Tiger Beer 330ml",
     "brand": "Tiger", "country_of_origin": "Singapore", "category": "Alcohol / Beer",
     "unit": "can", "price_usd": 1.05, "stock": 192, "stock_warning": 24,
     "pieces_per_package": 24, "pack_name": "case", "specs": "330ml x24/case",
     "is_featured": False, "is_active": True, "sort_order": 2, "description": "International premium beer."},
]

created = failed = 0
for p in PRODUCTS:
    r = requests.post(f"{BASE}/products", headers=H, json=p, verify=False)
    if r.status_code in (200, 201):
        created += 1; print(f"  OK {p['name']}")
    else:
        failed += 1; print(f"  FAIL {p['name']}: {r.text[:100]}")
print(f"  Created {created}, failed {failed}")

# 7. Create announcements
print("\n=== Create announcements ===")
ANNOUNCEMENTS = [
    {"type": "notice", "is_active": True, "sort_order": 1,
     "content_zh": "欢迎使用 KhmErai 批发商城！本平台专为柬埔寨批发商提供一站式采购服务，支持Telegram一键登录下单。如有疑问请联系客服。",
     "content_en": "Welcome to KhmErai Wholesale! One-stop procurement for Cambodia merchants. Supports Telegram login. Contact us for help."},
    {"type": "notice", "is_active": True, "sort_order": 2,
     "content_zh": "配送说明：金边市区订单满$50免运费；省份配送运费按距离计算；下单后1-2工作日发货。",
     "content_en": "Delivery: Free shipping over $50 in Phnom Penh. Provinces by distance. Ships in 1-2 business days."},
    {"type": "notice", "is_active": True, "sort_order": 3,
     "content_zh": "付款方式：现金(COD)、银行转账、ABA/Wing/TrueMoney、月结账期（需申请）。",
     "content_en": "Payment: COD, Bank Transfer, ABA/Wing/TrueMoney, Monthly Credit (application required)."},
    {"type": "contact", "is_active": True, "sort_order": 1,
     "content_zh": "联系我们：Telegram @testshopmouy_bot | 电话 +855 85 832 567 | 工作时间 周一至周六 8:00-18:00 | 地址：金边市，柬埔寨",
     "content_en": "Contact: Telegram @testshopmouy_bot | Phone +855 85 832 567 | Mon-Sat 8:00-18:00 | Phnom Penh, Cambodia"},
    {"type": "about", "is_active": True, "sort_order": 1,
     "content_zh": "关于我们：KhmErai是专注柬埔寨市场的批发平台，与多家知名品牌合作，为零售商和批发商提供优质实惠商品。",
     "content_en": "About: KhmErai is a wholesale platform for Cambodia, partnering with top brands to offer quality products at competitive prices."},
]
for a in ANNOUNCEMENTS:
    r = requests.post(f"{BASE}/announcements", headers=H, json=a, verify=False)
    print(f"  {'OK' if r.status_code in (200,201) else 'FAIL'} [{a['type']}]")

print("\n=== Done! Visit https://khmerai.cn ===")
