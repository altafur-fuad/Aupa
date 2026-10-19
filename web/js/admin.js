// ============================================================
// AUPA RESTAURANT PARTNER & ADMIN CONTROL PORTAL
// Order acceptance, Menu items, Store open/close, Revenue stats
// ============================================================

window.AupaAdmin = {
  isStoreOpen: true,
  orders: [],

  init() {
    this.orders = [...window.AupaData.partnerOrders];
    this.renderOrders();
    this.renderMenuItems();
    this.updateStats();
  },

  toggleStoreStatus() {
    this.isStoreOpen = !this.isStoreOpen;
    const badge = document.getElementById('storeStatusBadge');
    const btn = document.getElementById('storeToggleBtn');

    if (badge) {
      badge.textContent = this.isStoreOpen ? 'Accepting Orders (Online)' : 'Temporarily Closed (Offline)';
      badge.className = this.isStoreOpen ? 'badge-status badge-success' : 'badge-status badge-danger';
    }
    if (btn) {
      btn.textContent = this.isStoreOpen ? 'Close Restaurant' : 'Open Restaurant';
    }

    AupaCart.showToast(this.isStoreOpen ? 'Store is now ONLINE and receiving orders' : 'Store is now OFFLINE');
  },

  updateOrderStatus(orderId, newStatus) {
    const order = this.orders.find(o => o.id === orderId);
    if (!order) return;

    order.status = newStatus;
    this.renderOrders();
    this.updateStats();
    AupaCart.showToast(`Order ${orderId} marked as ${newStatus.toUpperCase()}`);
  },

  addNewMenuItem(name, price, tag, desc) {
    if (!name || !price) {
      alert('Please provide item name and price.');
      return;
    }

    const newItem = {
      id: 'm-' + Date.now(),
      name: name,
      price: parseInt(price, 10),
      tag: tag || 'Special',
      desc: desc || 'Freshly prepared specialty dish.'
    };

    // Add to first restaurant for demo
    window.AupaData.restaurants[0].menu.unshift(newItem);
    this.renderMenuItems();
    window.AupaApp.renderRestaurants();
    AupaCart.showToast(`Added "${name}" to menu!`);
  },

  renderOrders() {
    const container = document.getElementById('adminOrdersTableBody');
    if (!container) return;

    if (this.orders.length === 0) {
      container.innerHTML = '<tr><td colspan="6" style="text-align:center; padding:24px; color:#94A3B8;">No incoming orders right now</td></tr>';
      return;
    }

    let html = '';
    this.orders.forEach(order => {
      let statusBadge = '';
      let actionBtns = '';

      if (order.status === 'pending') {
        statusBadge = '<span style="background:#FEF3C7; color:#B45309; padding:3px 8px; border-radius:4px; font-weight:700; font-size:0.78rem;">PENDING</span>';
        actionBtns = `
          <button class="btn btn-sm btn-primary" onclick="AupaAdmin.updateOrderStatus('${order.id}', 'preparing')">Accept</button>
          <button class="btn btn-sm btn-outline" style="color:#EF4444;" onclick="AupaAdmin.updateOrderStatus('${order.id}', 'cancelled')">Reject</button>
        `;
      } else if (order.status === 'preparing') {
        statusBadge = '<span style="background:#DBEAFE; color:#1D4ED8; padding:3px 8px; border-radius:4px; font-weight:700; font-size:0.78rem;">PREPARING</span>';
        actionBtns = `
          <button class="btn btn-sm btn-navy" onclick="AupaAdmin.updateOrderStatus('${order.id}', 'ready')">Mark Ready</button>
        `;
      } else if (order.status === 'ready') {
        statusBadge = '<span style="background:#E0E7FF; color:#4338CA; padding:3px 8px; border-radius:4px; font-weight:700; font-size:0.78rem;">RIDER ASSIGNED</span>';
        actionBtns = `
          <button class="btn btn-sm btn-primary" onclick="AupaAdmin.updateOrderStatus('${order.id}', 'delivered')">Mark Delivered</button>
        `;
      } else if (order.status === 'delivered') {
        statusBadge = '<span style="background:#D1FAE5; color:#065F46; padding:3px 8px; border-radius:4px; font-weight:700; font-size:0.78rem;">DELIVERED</span>';
        actionBtns = `<span style="color:#64748B; font-size:0.85rem;">Completed</span>`;
      }

      const itemSummary = order.items.map(i => `${i.qty}x ${i.name}`).join(', ');

      html += `
        <tr style="border-bottom:1px solid #E2E8F0;">
          <td style="padding:14px; font-weight:700;">#${order.id}</td>
          <td style="padding:14px;"><strong>${order.customerName}</strong><br><small style="color:#64748B;">${order.customerPhone}</small></td>
          <td style="padding:14px; font-size:0.88rem; max-width:240px;">${itemSummary}</td>
          <td style="padding:14px; font-weight:700;">৳${order.total}</td>
          <td style="padding:14px;">${statusBadge}</td>
          <td style="padding:14px; display:flex; gap:6px;">${actionBtns}</td>
        </tr>
      `;
    });

    container.innerHTML = html;
  },

  renderMenuItems() {
    const container = document.getElementById('adminMenuItemsList');
    if (!container) return;

    const myRestaurant = window.AupaData.restaurants[0];
    let html = '';
    myRestaurant.menu.forEach(item => {
      html += `
        <div style="display:flex; align-items:center; justify-content:space-between; padding:12px 16px; border:1px solid #E2E8F0; border-radius:10px; margin-bottom:8px; background:white;">
          <div>
            <strong>${item.name}</strong> <span style="background:#F1F5F9; font-size:0.75rem; padding:2px 6px; border-radius:4px; font-weight:600;">${item.tag}</span>
            <div style="font-size:0.82rem; color:#64748B;">৳${item.price} • ${item.desc.substring(0, 45)}...</div>
          </div>
          <div>
            <span style="color:#10B981; font-weight:700; font-size:0.85rem;">In Stock</span>
          </div>
        </div>
      `;
    });
    container.innerHTML = html;
  },

  updateStats() {
    const totalRev = this.orders.reduce((sum, o) => sum + (o.status !== 'cancelled' ? o.total : 0), 0);
    const completedCount = this.orders.filter(o => o.status === 'delivered').length;
    const activeCount = this.orders.filter(o => o.status === 'pending' || o.status === 'preparing' || o.status === 'ready').length;

    const revEl = document.getElementById('statTotalRevenue');
    const compEl = document.getElementById('statCompletedOrders');
    const actEl = document.getElementById('statActiveOrders');

    if (revEl) revEl.textContent = `৳${totalRev.toLocaleString()}`;
    if (compEl) compEl.textContent = completedCount;
    if (actEl) actEl.textContent = activeCount;
  }
};
