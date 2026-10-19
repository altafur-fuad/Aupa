// ============================================================
// AUPA MAIN WEB CLIENT APPLICATION
// ============================================================

window.AupaApp = {
  currentCategory: 'all',
  searchQuery: '',
  activeRestaurant: null,

  init() {
    this.renderCategories();
    this.renderRestaurants();
    this.setupEventListeners();
    AupaCart.init();
    AupaAdmin.init();
  },

  setupEventListeners() {
    // Search input
    const searchInput = document.getElementById('heroSearchInput');
    if (searchInput) {
      searchInput.addEventListener('input', (e) => {
        this.searchQuery = e.target.value.toLowerCase().trim();
        this.renderRestaurants();
      });
      searchInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
          this.scrollToRestaurants();
        }
      });
    }

    // Close modal on escape
    window.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        this.closeAllModals();
        this.closeCart();
      }
    });
  },

  renderCategories() {
    const container = document.getElementById('categoryScroll');
    if (!container) return;

    let html = '';
    window.AupaData.categories.forEach(cat => {
      const activeClass = cat.id === this.currentCategory ? 'active' : '';
      html += `
        <button class="cat-chip ${activeClass}" onclick="AupaApp.filterCategory('${cat.id}')">
          <span class="icon">${cat.icon}</span>
          <span>${cat.name}</span>
        </button>
      `;
    });
    container.innerHTML = html;
  },

  filterCategory(catId) {
    this.currentCategory = catId;
    this.renderCategories();
    this.renderRestaurants();
    this.scrollToRestaurants();
  },

  setQuickLocation(locName) {
    const input = document.getElementById('heroSearchInput');
    if (input) {
      input.value = locName;
      this.searchQuery = locName.toLowerCase();
      this.renderRestaurants();
      this.scrollToRestaurants();
      AupaCart.showToast(`Delivering to ${locName}`);
    }
  },

  renderRestaurants() {
    const grid = document.getElementById('restaurantGrid');
    const countEl = document.getElementById('restaurantCountText');
    if (!grid) return;

    const filtered = window.AupaData.restaurants.filter(r => {
      const matchCat = this.currentCategory === 'all' || r.category === this.currentCategory;
      const matchSearch = !this.searchQuery ||
        r.name.toLowerCase().includes(this.searchQuery) ||
        r.cuisines.some(c => c.toLowerCase().includes(this.searchQuery));
      return matchCat && matchSearch;
    });

    if (countEl) {
      countEl.textContent = `Showing ${filtered.length} curated restaurants nearby`;
    }

    if (filtered.length === 0) {
      grid.innerHTML = `
        <div style="grid-column: 1 / -1; text-align: center; padding: 60px 20px; background: white; border-radius: 16px; border: 1px solid var(--border-light);">
          <div style="font-size: 3rem; margin-bottom: 12px;">🔍</div>
          <h3 style="font-size: 1.3rem; margin-bottom: 8px;">No restaurants found</h3>
          <p style="color: var(--text-muted); max-width: 400px; margin: 0 auto 20px;">We couldn't find any spots matching "${this.searchQuery}". Try selecting another category or location.</p>
          <button class="btn btn-primary btn-sm" onclick="AupaApp.resetFilters()">Clear Filters</button>
        </div>
      `;
      return;
    }

    let html = '';
    filtered.forEach(r => {
      const freeBadge = r.isFreeDelivery ? '<span>• Free Delivery</span>' : `<span>• ৳${r.deliveryFee} fee</span>`;
      html += `
        <div class="restaurant-card" onclick="AupaApp.openRestaurantMenu('${r.id}')">
          <div class="card-media">
            <img src="${r.image}" alt="${r.name}" loading="lazy">
            <div class="card-badge-top">${r.badge}</div>
            <div class="card-badge-delivery">
              <span>⏱️ ${r.deliveryTime}</span>
            </div>
          </div>
          <div class="card-body">
            <div class="card-header-line">
              <h3 class="card-title">${r.name}</h3>
              <div class="card-rating">
                <span>★</span>
                <span>${r.rating}</span>
              </div>
            </div>
            <p class="card-cuisines">${r.cuisines.join(' • ')}</p>
            <div class="card-footer-line">
              <span>Min order: <strong>৳${r.minOrder}</strong></span>
              ${freeBadge}
            </div>
          </div>
        </div>
      `;
    });

    grid.innerHTML = html;
  },

  resetFilters() {
    this.currentCategory = 'all';
    this.searchQuery = '';
    const input = document.getElementById('heroSearchInput');
    if (input) input.value = '';
    this.renderCategories();
    this.renderRestaurants();
  },

  scrollToRestaurants() {
    const el = document.getElementById('restaurantsSection');
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  },

  openRestaurantMenu(restaurantId) {
    const restaurant = window.AupaData.restaurants.find(r => r.id === restaurantId);
    if (!restaurant) return;

    this.activeRestaurant = restaurant;
    const modal = document.getElementById('menuModal');
    const titleEl = document.getElementById('menuModalTitle');
    const subEl = document.getElementById('menuModalSubtitle');
    const itemsEl = document.getElementById('menuModalItems');

    if (titleEl) titleEl.textContent = restaurant.name;
    if (subEl) subEl.textContent = `${restaurant.cuisines.join(' • ')} • ★ ${restaurant.rating} (${restaurant.reviewsCount}+ ratings) • ⏱️ ${restaurant.deliveryTime}`;

    let html = '';
    restaurant.menu.forEach(item => {
      html += `
        <div style="display:flex; justify-content:space-between; align-items:flex-start; padding:18px 0; border-bottom:1px solid var(--border-light); gap:20px;">
          <div style="flex:1;">
            <div style="display:flex; align-items:center; gap:8px; margin-bottom:4px;">
              <h4 style="font-size:1.05rem; font-weight:700;">${item.name}</h4>
              <span style="background:var(--primary-light); color:var(--primary); font-size:0.75rem; font-weight:700; padding:2px 8px; border-radius:9999px;">${item.tag}</span>
            </div>
            <p style="font-size:0.88rem; color:var(--text-muted); line-height:1.5; margin-bottom:8px;">${item.desc}</p>
            <strong style="font-size:1.05rem; color:var(--text-main);">৳${item.price}</strong>
          </div>
          <button class="btn btn-sm btn-primary" onclick="AupaApp.addToCartFromModal('${item.id}')">
            + Add to Cart
          </button>
        </div>
      `;
    });

    if (itemsEl) itemsEl.innerHTML = html;
    if (modal) modal.classList.add('active');
  },

  addToCartFromModal(itemId) {
    if (!this.activeRestaurant) return;
    const item = this.activeRestaurant.menu.find(m => m.id === itemId);
    if (!item) return;

    AupaCart.addItem(this.activeRestaurant, item);
  },

  openCart() {
    const drawer = document.getElementById('cartDrawer');
    const overlay = document.getElementById('cartOverlay');
    if (drawer) drawer.classList.add('open');
    if (overlay) overlay.classList.add('active');
  },

  closeCart() {
    const drawer = document.getElementById('cartDrawer');
    const overlay = document.getElementById('cartOverlay');
    if (drawer) drawer.classList.remove('open');
    if (overlay) overlay.classList.remove('active');
  },

  openCheckoutModal() {
    if (AupaCart.items.length === 0) {
      AupaCart.showToast('Your cart is empty', true);
      return;
    }
    this.closeCart();
    const modal = document.getElementById('checkoutModal');
    const totalEl = document.getElementById('checkoutTotalText');
    if (totalEl) totalEl.textContent = `৳${AupaCart.getTotal()}`;
    if (modal) modal.classList.add('active');
  },

  handleCheckoutSubmit(event) {
    event.preventDefault();
    const name = document.getElementById('checkoutName').value.trim();
    const phone = document.getElementById('checkoutPhone').value.trim();
    const address = document.getElementById('checkoutAddress').value.trim();
    const payment = document.querySelector('input[name="paymentMethod"]:checked')?.value || 'Cash on Delivery';

    if (!name || !phone || !address) {
      alert('Please fill out all required delivery fields.');
      return;
    }

    // Create live order
    const orderId = 'AUP-' + Math.floor(1000 + Math.random() * 9000);
    const newOrder = {
      id: orderId,
      customerName: name,
      customerPhone: phone,
      address: address,
      items: AupaCart.items.map(i => ({ name: i.name, qty: i.quantity, price: i.price })),
      total: AupaCart.getTotal(),
      paymentMethod: payment,
      status: 'pending',
      timestamp: 'Just now'
    };

    // Push into Admin live feed
    window.AupaData.partnerOrders.unshift(newOrder);
    AupaAdmin.orders = [...window.AupaData.partnerOrders];
    AupaAdmin.renderOrders();
    AupaAdmin.updateStats();

    // Clear user cart
    AupaCart.clearCart();
    this.closeAllModals();

    // Open live order tracking modal
    this.showOrderTracker(newOrder);
  },

  showOrderTracker(order) {
    const modal = document.getElementById('trackingModal');
    const idEl = document.getElementById('trackerOrderId');
    const totalEl = document.getElementById('trackerTotal');
    const addressEl = document.getElementById('trackerAddress');

    if (idEl) idEl.textContent = `#${order.id}`;
    if (totalEl) totalEl.textContent = `৳${order.total} (${order.paymentMethod})`;
    if (addressEl) addressEl.textContent = order.address;

    if (modal) modal.classList.add('active');
  },

  downloadApk() {
    AupaCart.showToast('Starting Aupa Android APK download...');
    
    // Trigger download of existing apk file
    const link = document.createElement('a');
    link.href = 'downloads/aupa.apk';
    link.download = 'aupa.apk';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);

    const progressEl = document.getElementById('apkDownloadProgress');
    if (progressEl) {
      progressEl.style.display = 'block';
      let progress = 0;
      const interval = setInterval(() => {
        progress += 25;
        progressEl.value = progress;
        if (progress >= 100) {
          clearInterval(interval);
          setTimeout(() => {
            progressEl.style.display = 'none';
            progressEl.value = 0;
            AupaCart.showToast('Aupa APK downloaded successfully! Tap to install on Android.');
          }, 600);
        }
      }, 300);
    }
  },

  switchView(viewName) {
    const storefront = document.getElementById('storefrontView');
    const portal = document.getElementById('portalView');
    const navStorefrontBtn = document.getElementById('navStorefrontLink');
    const navPartnerBtn = document.getElementById('navPartnerLink');

    if (viewName === 'partner') {
      if (storefront) storefront.style.display = 'none';
      if (portal) portal.classList.add('active');
      if (navStorefrontBtn) navStorefrontBtn.classList.remove('active');
      if (navPartnerBtn) navPartnerBtn.classList.add('active');
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } else {
      if (storefront) storefront.style.display = 'block';
      if (portal) portal.classList.remove('active');
      if (navStorefrontBtn) navStorefrontBtn.classList.add('active');
      if (navPartnerBtn) navPartnerBtn.classList.remove('active');
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  },

  openAuthModal(type = 'signin') {
    const modal = document.getElementById('authModal');
    const titleEl = document.getElementById('authModalTitle');
    if (titleEl) {
      titleEl.textContent = type === 'signup' ? 'Create an Aupa Account' : 'Welcome Back to Aupa';
    }
    if (modal) modal.classList.add('active');
  },

  openPartnerApplicationModal(role) {
    const modal = document.getElementById('partnerModal');
    const title = document.getElementById('partnerModalTitle');
    const subtitle = document.getElementById('partnerModalSubtitle');

    if (role === 'rider') {
      if (title) title.textContent = 'Sign up to Deliver with Aupa';
      if (subtitle) subtitle.textContent = 'Earn competitive per-delivery payouts, flexible hours, and instant tips.';
    } else if (role === 'restaurant') {
      if (title) title.textContent = 'Partner Your Restaurant with Aupa';
      if (subtitle) subtitle.textContent = 'Expand your customer reach across Dhaka with 30-min reliable delivery.';
    } else {
      if (title) title.textContent = 'Join the Aupa Team';
      if (subtitle) subtitle.textContent = 'Build Bangladesh’s fastest and most reliable food delivery network.';
    }

    if (modal) modal.classList.add('active');
  },

  closeAllModals() {
    document.querySelectorAll('.modal-overlay').forEach(m => m.classList.remove('active'));
  }
};

// Initialize on DOM Ready
document.addEventListener('DOMContentLoaded', () => {
  window.AupaApp.init();
});
