// ============================================================
// AUPA SHOPPING CART & ORDER LIFECYCLE CONTROLLER
// ============================================================

window.AupaCart = {
  items: [],
  currentRestaurantId: null,
  currentRestaurantName: '',
  deliveryFee: 40,
  promoDiscount: 0,
  activeOrder: null,

  init() {
    this.loadFromStorage();
    this.updateUI();
  },

  loadFromStorage() {
    try {
      const stored = localStorage.getItem('aupa_web_cart');
      if (stored) {
        const parsed = JSON.parse(stored);
        this.items = parsed.items || [];
        this.currentRestaurantId = parsed.currentRestaurantId || null;
        this.currentRestaurantName = parsed.currentRestaurantName || '';
      }
    } catch (e) {
      this.items = [];
    }
  },

  saveToStorage() {
    try {
      localStorage.setItem('aupa_web_cart', JSON.stringify({
        items: this.items,
        currentRestaurantId: this.currentRestaurantId,
        currentRestaurantName: this.currentRestaurantName
      }));
    } catch (e) {}
  },

  addItem(restaurant, menuItem) {
    if (this.currentRestaurantId && this.currentRestaurantId !== restaurant.id && this.items.length > 0) {
      if (!confirm(`Your cart already contains items from "${this.currentRestaurantName}". Would you like to clear your cart and start a new order from "${restaurant.name}"?`)) {
        return;
      }
      this.items = [];
    }

    this.currentRestaurantId = restaurant.id;
    this.currentRestaurantName = restaurant.name;
    this.deliveryFee = restaurant.isFreeDelivery ? 0 : restaurant.deliveryFee;

    const existing = this.items.find(i => i.id === menuItem.id);
    if (existing) {
      existing.quantity += 1;
    } else {
      this.items.push({
        id: menuItem.id,
        name: menuItem.name,
        price: menuItem.price,
        quantity: 1,
        restaurantId: restaurant.id
      });
    }

    this.saveToStorage();
    this.updateUI();
    this.showToast(`Added ${menuItem.name} to cart!`);
  },

  updateQuantity(itemId, delta) {
    const item = this.items.find(i => i.id === itemId);
    if (!item) return;

    item.quantity += delta;
    if (item.quantity <= 0) {
      this.items = this.items.filter(i => i.id !== itemId);
    }

    if (this.items.length === 0) {
      this.currentRestaurantId = null;
      this.currentRestaurantName = '';
      this.promoDiscount = 0;
    }

    this.saveToStorage();
    this.updateUI();
  },

  clearCart() {
    this.items = [];
    this.currentRestaurantId = null;
    this.currentRestaurantName = '';
    this.promoDiscount = 0;
    this.saveToStorage();
    this.updateUI();
  },

  getSubtotal() {
    return this.items.reduce((sum, item) => sum + (item.price * item.quantity), 0);
  },

  getVat() {
    return Math.round(this.getSubtotal() * 0.05); // 5% VAT
  },

  getTotal() {
    const sub = this.getSubtotal();
    if (sub === 0) return 0;
    const total = sub + this.deliveryFee + this.getVat() - this.promoDiscount;
    return Math.max(total, 0);
  },

  applyPromoCode(code) {
    const clean = (code || '').trim().toUpperCase();
    if (clean === 'AUPA50' || clean === 'FIRSTBITE') {
      this.promoDiscount = 50;
      this.updateUI();
      this.showToast('Promo code applied! ৳50 discount applied.');
      return true;
    } else {
      this.showToast('Invalid promo code. Try "AUPA50"', true);
      return false;
    }
  },

  updateUI() {
    // 1. Badge count
    const totalQty = this.items.reduce((sum, item) => sum + item.quantity, 0);
    const countBadges = document.querySelectorAll('.cart-count');
    countBadges.forEach(b => {
      b.textContent = totalQty;
      b.style.display = totalQty > 0 ? 'flex' : 'none';
    });

    // 2. Render items inside drawer
    const listEl = document.getElementById('cartItemsList');
    const emptyEl = document.getElementById('cartEmptyState');
    const summaryEl = document.getElementById('cartFooterSummary');

    if (!listEl) return;

    if (this.items.length === 0) {
      listEl.innerHTML = '';
      if (emptyEl) emptyEl.style.display = 'block';
      if (summaryEl) summaryEl.style.display = 'none';
      return;
    }

    if (emptyEl) emptyEl.style.display = 'none';
    if (summaryEl) summaryEl.style.display = 'block';

    let html = '';
    this.items.forEach(item => {
      html += `
        <div class="cart-item">
          <div class="cart-item-info">
            <h4>${item.name}</h4>
            <p>৳${item.price} each</p>
          </div>
          <div class="qty-control">
            <button class="qty-btn" onclick="AupaCart.updateQuantity('${item.id}', -1)">-</button>
            <span style="font-weight:700; font-size:0.9rem;">${item.quantity}</span>
            <button class="qty-btn" onclick="AupaCart.updateQuantity('${item.id}', 1)">+</button>
          </div>
        </div>
      `;
    });
    listEl.innerHTML = html;

    // Subtotal and Totals
    const subtotalEl = document.getElementById('cartSubtotal');
    const vatEl = document.getElementById('cartVat');
    const deliveryEl = document.getElementById('cartDelivery');
    const totalEl = document.getElementById('cartTotal');

    if (subtotalEl) subtotalEl.textContent = `৳${this.getSubtotal()}`;
    if (vatEl) vatEl.textContent = `৳${this.getVat()}`;
    if (deliveryEl) deliveryEl.textContent = this.deliveryFee === 0 ? 'Free' : `৳${this.deliveryFee}`;
    if (totalEl) totalEl.textContent = `৳${this.getTotal()}`;
  },

  showToast(msg, isError = false) {
    let toast = document.getElementById('aupaToast');
    if (!toast) {
      toast = document.createElement('div');
      toast.id = 'aupaToast';
      toast.style.cssText = `
        position: fixed;
        bottom: 24px;
        right: 24px;
        background: #151F32;
        color: white;
        padding: 12px 24px;
        border-radius: 9999px;
        box-shadow: 0 10px 25px rgba(0,0,0,0.2);
        z-index: 9999;
        font-weight: 600;
        font-size: 0.92rem;
        transition: transform 0.3s ease, opacity 0.3s ease;
        opacity: 0;
        transform: translateY(20px);
        pointer-events: none;
      `;
      document.body.appendChild(toast);
    }

    toast.textContent = msg;
    toast.style.background = isError ? '#EF4444' : '#151F32';
    toast.style.opacity = '1';
    toast.style.transform = 'translateY(0)';

    setTimeout(() => {
      toast.style.opacity = '0';
      toast.style.transform = 'translateY(20px)';
    }, 2800);
  }
};
