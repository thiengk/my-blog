<script>
  /**
   * MealDashboard Component
   * Main dashboard showing next payer, stats, and recent payment history.
   */
  import { mealFetch } from '../lib/meal-api';

  /** @type {{ breakfast: any, lunch: any } | null} */
  let nextPayer = $state(null);
  /** @type {any[]} */
  let stats = $state([]);
  /** @type {any[]} */
  let payments = $state([]);
  /** @type {any[]} */
  let members = $state([]);
  let loading = $state(true);
  let error = $state('');
  let confirmingMeal = $state('');
  let overrideMode = $state('');
  let selectedOverrideMember = $state(0);

  async function loadData() {
    loading = true;
    error = '';
    try {
      const [npRes, statsRes, paymentsRes, membersRes] = await Promise.all([
        mealFetch('/api/meals/next-payer'),
        mealFetch('/api/meals/stats'),
        mealFetch('/api/meals/payments?limit=10&offset=0'),
        mealFetch('/api/meals/members'),
      ]);
      nextPayer = npRes;
      stats = statsRes.stats || [];
      payments = paymentsRes.payments || [];
      members = membersRes.members || [];
    } catch (e) {
      error = 'Không thể tải dữ liệu. Vui lòng thử lại.';
    }
    loading = false;
  }

  $effect(() => {
    loadData();
  });

  /**
   * Confirm payment for the next payer.
   * @param {string} mealType
   */
  async function confirmPayment(mealType) {
    const payer = mealType === 'breakfast' ? nextPayer?.breakfast : nextPayer?.lunch;
    if (!payer) return;

    confirmingMeal = mealType;
    try {
      await mealFetch('/api/meals/payments', {
        method: 'POST',
        body: JSON.stringify({
          member_id: payer.member_id,
          meal_type: mealType,
          date: new Date().toISOString().split('T')[0],
        }),
      });
      await loadData();
    } catch (e) {
      error = 'Không thể ghi nhận thanh toán.';
    }
    confirmingMeal = '';
  }

  /**
   * Record payment for a different member (override).
   * @param {string} mealType
   */
  async function confirmOverride(mealType) {
    if (!selectedOverrideMember) return;

    confirmingMeal = mealType;
    try {
      await mealFetch('/api/meals/payments', {
        method: 'POST',
        body: JSON.stringify({
          member_id: selectedOverrideMember,
          meal_type: mealType,
          date: new Date().toISOString().split('T')[0],
        }),
      });
      overrideMode = '';
      selectedOverrideMember = 0;
      await loadData();
    } catch (e) {
      error = 'Không thể ghi nhận thanh toán.';
    }
    confirmingMeal = '';
  }

  /**
   * Undo the most recent payment.
   * @param {number} paymentId
   */
  async function undoPayment(paymentId) {
    try {
      await mealFetch(`/api/meals/payments/${paymentId}`, { method: 'DELETE' });
      await loadData();
    } catch (e) {
      error = 'Không thể hoàn tác. Có thể đã quá 24 giờ.';
    }
  }

  /**
   * Check if a payment can be undone (within 24h).
   * @param {string} createdAt
   * @returns {boolean}
   */
  function canUndo(createdAt) {
    const created = new Date(createdAt);
    const now = new Date();
    return (now.getTime() - created.getTime()) < 24 * 60 * 60 * 1000;
  }

  function getMealLabel(type) {
    return type === 'breakfast' ? '🌅 Sáng' : '☀️ Trưa';
  }
</script>

{#if loading}
  <div class="meal-loading">Đang tải...</div>
{:else if error}
  <div class="meal-error" role="alert">{error}
    <button onclick={loadData}>Thử lại</button>
  </div>
{:else}
  <div class="meal-dashboard">
    <!-- Next Payer Cards -->
    <section class="next-payer-section">
      <h2>Người trả tiếp theo</h2>
      <div class="payer-cards">
        {#each ['breakfast', 'lunch'] as mealType}
          {@const payer = mealType === 'breakfast' ? nextPayer?.breakfast : nextPayer?.lunch}
          <div class="payer-card">
            <div class="payer-meal-label">{getMealLabel(mealType)}</div>
            {#if payer}
              <div class="payer-name">{payer.member_name}</div>
              <div class="payer-count">Đã trả: {payer.payment_count} lần</div>
              <div class="payer-actions">
                <button
                  class="btn-confirm"
                  onclick={() => confirmPayment(mealType)}
                  disabled={confirmingMeal === mealType}
                >
                  {confirmingMeal === mealType ? '...' : '✓ Xác nhận đã trả'}
                </button>
                <button
                  class="btn-override"
                  onclick={() => { overrideMode = mealType; selectedOverrideMember = 0; }}
                >
                  Người khác trả
                </button>
              </div>
              {#if overrideMode === mealType}
                <div class="override-form">
                  <select bind:value={selectedOverrideMember}>
                    <option value={0}>-- Chọn người --</option>
                    {#each members.filter(m => m.is_active) as member}
                      <option value={member.id}>{member.name}</option>
                    {/each}
                  </select>
                  <button
                    class="btn-confirm-sm"
                    onclick={() => confirmOverride(mealType)}
                    disabled={!selectedOverrideMember}
                  >
                    Xác nhận
                  </button>
                  <button class="btn-cancel-sm" onclick={() => { overrideMode = ''; }}>Hủy</button>
                </div>
              {/if}
            {:else}
              <div class="payer-empty">Chưa có ai tham gia</div>
            {/if}
          </div>
        {/each}
      </div>
    </section>

    <!-- Stats Table -->
    <section class="stats-section">
      <h2>Thống kê thanh toán</h2>
      <div class="table-wrapper">
        <table>
          <thead>
            <tr>
              <th>Thành viên</th>
              <th>Sáng</th>
              <th>Trưa</th>
              <th>Tổng</th>
            </tr>
          </thead>
          <tbody>
            {#each stats as stat}
              <tr>
                <td>{stat.member_name}</td>
                <td>{stat.breakfast_count}</td>
                <td>{stat.lunch_count}</td>
                <td><strong>{stat.total_count}</strong></td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </section>

    <!-- Recent Payments -->
    <section class="history-section">
      <h2>Lịch sử gần đây</h2>
      {#if payments.length === 0}
        <p class="empty-state">Chưa có lịch sử thanh toán</p>
      {:else}
        <ul class="payment-list">
          {#each payments as payment, i}
            <li class="payment-item">
              <span class="payment-info">
                <strong>{payment.member_name}</strong> trả {getMealLabel(payment.meal_type)} — {payment.payment_date}
              </span>
              {#if i === 0 && canUndo(payment.created_at)}
                <button class="btn-undo" onclick={() => undoPayment(payment.id)}>Hoàn tác</button>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  </div>
{/if}

<style>
  .meal-loading, .meal-error {
    text-align: center;
    padding: 2rem;
    color: var(--text-muted, #718096);
  }

  .meal-error {
    color: #ef4444;
  }

  .meal-error button {
    margin-top: 0.5rem;
    padding: 0.5rem 1rem;
    border: 1px solid #ef4444;
    border-radius: 0.375rem;
    background: transparent;
    color: #ef4444;
    cursor: pointer;
  }

  .meal-dashboard {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  h2 {
    font-size: 1.25rem;
    margin: 0 0 1rem;
    color: var(--text-color, #1a202c);
  }

  .payer-cards {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  @media (max-width: 640px) {
    .payer-cards {
      grid-template-columns: 1fr;
    }
  }

  .payer-card {
    padding: 1.5rem;
    border: 1px solid var(--border-color, #e2e8f0);
    border-radius: 0.75rem;
    background: var(--card-bg, #ffffff);
  }

  .payer-meal-label {
    font-size: 0.875rem;
    color: var(--text-muted, #718096);
    margin-bottom: 0.5rem;
  }

  .payer-name {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-color, #1a202c);
  }

  .payer-count {
    font-size: 0.875rem;
    color: var(--text-muted, #718096);
    margin-top: 0.25rem;
  }

  .payer-actions {
    margin-top: 1rem;
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .payer-empty {
    color: var(--text-muted, #718096);
    font-style: italic;
  }

  .btn-confirm {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 0.375rem;
    background: #10b981;
    color: #ffffff;
    font-size: 0.875rem;
    cursor: pointer;
    transition: background 0.2s;
  }

  .btn-confirm:hover:not(:disabled) { background: #059669; }
  .btn-confirm:disabled { opacity: 0.5; cursor: not-allowed; }

  .btn-override {
    padding: 0.5rem 1rem;
    border: 1px solid var(--border-color, #e2e8f0);
    border-radius: 0.375rem;
    background: transparent;
    color: var(--text-color, #4a5568);
    font-size: 0.875rem;
    cursor: pointer;
  }

  .btn-override:hover { background: var(--hover-bg, #f7fafc); }

  .override-form {
    margin-top: 0.75rem;
    display: flex;
    gap: 0.5rem;
    align-items: center;
    flex-wrap: wrap;
  }

  .override-form select {
    padding: 0.375rem 0.5rem;
    border: 1px solid var(--border-color, #e2e8f0);
    border-radius: 0.375rem;
    font-size: 0.875rem;
    background: var(--input-bg, #ffffff);
    color: var(--text-color, #1a202c);
  }

  .btn-confirm-sm {
    padding: 0.375rem 0.75rem;
    border: none;
    border-radius: 0.375rem;
    background: #10b981;
    color: #ffffff;
    font-size: 0.8rem;
    cursor: pointer;
  }

  .btn-confirm-sm:disabled { opacity: 0.5; cursor: not-allowed; }

  .btn-cancel-sm {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--border-color, #e2e8f0);
    border-radius: 0.375rem;
    background: transparent;
    color: var(--text-muted, #718096);
    font-size: 0.8rem;
    cursor: pointer;
  }

  .table-wrapper {
    overflow-x: auto;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  th, td {
    padding: 0.75rem;
    text-align: left;
    border-bottom: 1px solid var(--border-color, #e2e8f0);
  }

  th {
    font-weight: 600;
    color: var(--text-muted, #718096);
    font-size: 0.8rem;
    text-transform: uppercase;
  }

  td {
    color: var(--text-color, #1a202c);
  }

  .payment-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .payment-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 0;
    border-bottom: 1px solid var(--border-color, #e2e8f0);
    font-size: 0.875rem;
  }

  .payment-info {
    color: var(--text-color, #1a202c);
  }

  .btn-undo {
    padding: 0.25rem 0.5rem;
    border: 1px solid #f59e0b;
    border-radius: 0.25rem;
    background: transparent;
    color: #f59e0b;
    font-size: 0.75rem;
    cursor: pointer;
  }

  .btn-undo:hover { background: rgba(245, 158, 11, 0.1); }

  .empty-state {
    color: var(--text-muted, #718096);
    font-style: italic;
  }

  :global(.dark) .payer-card {
    background: var(--card-bg, #1e293b);
    border-color: var(--border-color, #334155);
  }

  :global(.dark) .payer-name,
  :global(.dark) h2,
  :global(.dark) td,
  :global(.dark) .payment-info {
    color: var(--text-color, #f1f5f9);
  }

  :global(.dark) .override-form select {
    background: var(--input-bg, #0f172a);
    border-color: var(--border-color, #334155);
    color: var(--text-color, #f1f5f9);
  }
</style>
