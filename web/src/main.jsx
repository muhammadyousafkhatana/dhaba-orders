import React, { useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';
import './style.css';

function App() {
  const [menu, setMenu] = useState([]);
  const [orders, setOrders] = useState([]);
  const [customer, setCustomer] = useState('');
  const [selected, setSelected] = useState({});

  async function load() {
    try {
      const response = await fetch('/api/menu');
      if (response.ok) setMenu(await response.json());
    } catch (e) { console.error(e); }
  }

  async function placeOrder(e) {
    e.preventDefault();
    const items = Object.entries(selected)
      .filter(([, quantity]) => quantity > 0)
      .map(([id, quantity]) => ({menu_item_id: Number(id), quantity}));
    if (!customer || !items.length) return alert('Enter your name and choose food');

    const response = await fetch('/api/orders', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({customer_name: customer, items})
    });
    const data = await response.json();
    if (!response.ok) return alert(data.error || 'Order failed');
    alert(`Order #${data.id} placed successfully! Total: ${data.total_amount}`);
    setSelected({});
    setCustomer('');
    setOrders(old => [data, ...old]);
  }

  useEffect(() => { load(); }, []);

  return <main>
    <header><h1>🍛 Dhaba Orders</h1><p>Fresh food. Fast service.</p></header>
    <section><h2>Our Menu</h2>
      {menu.length ? menu.map(item => <article key={item.id}>
        <div><strong>{item.name}</strong><p>{item.description}</p>
        <b>Rs. {item.price}</b></div>
        <input type="number" min="0" value={selected[item.id] || 0}
          onChange={e => setSelected({...selected, [item.id]: Number(e.target.value)})}/>
      </article>) : <p>Menu is loading or API is unavailable.</p>}
    </section>
    <form onSubmit={placeOrder}>
      <h2>Place your order</h2>
      <input placeholder="Your name" value={customer}
        onChange={e => setCustomer(e.target.value)} required/>
      <button type="submit">Place Order</button>
    </form>
    <section><h2>Recent Orders</h2>
      {orders.map((o, i) => <p key={i}>Order #{o.id} — {o.status}</p>)}
    </section>
  </main>;
}

createRoot(document.getElementById('root')).render(<App />);
