<script>
  /**
   * MemberManager Component
   * CRUD for team members and meal participation configuration.
   */
  import { mealFetch } from '../lib/meal-api';

  /** @type {any[]} */
  let members = $state([]);
  /** @type {any[]} */
  let participations = $state([]);
  let newName = $state('');
  let loading = $state(true);
  let error = $state('');
  let editingId = $state(0);
  let editName = $state('');

  async function loadData() {
    loading = true;
    error = '';
    try {
      const [membersRes, partRes] = await Promise.all([
        mealFetch('/api/meals/members'),
        mealFetch('/api/meals/participations'),
      ]);
      members = membersRes.members || [];
      participations = partRes.participations || [];
    } catch (e) {
      error = 'Không thể tải dữ liệu.';
    }
    loading = false;
  }

  $effect(() => { loadData(); });

  /**
   * Get participation for a member and meal type.
   */
  function getParticipation(memberId, mealType) {
    return participations.find(p => p.member_id === memberId && p.meal_type === mealType);
  }

  function isParticipating(memberId, mealType) {
    const p = getParticipation(memberId, mealType);
    return p ? p.is_participating : true;
  }

  async function addMember() {
    if (!newName.trim()) return;
    error = '';
    try {
      await mealFetch('/api/meals/members', {
        method: 'POST',
        body: JSON.stringify({ name: newName.trim() }),
      });
      newName = '';
      await loadData();
    } catch (e) {
      error = e.message || 'Không thể thêm thành viên.';
    }
  }

  async function toggleActive(member) {
    try {
      await mealFetch(`/api/meals/members/${member.id}`, {
        method: 'PUT',
        body: JSON.stringify({ name: member.name, is_active: !member.is_active }),
      });
      await loadData();
    } catch (e) {
      error = 'Không thể cập nhật.';
    }
  }

  async function saveName(member) {
    if (!editName.trim()) return;
    try {
      await mealFetch(`/api/meals/members/${member.id}`, {
        method: 'PUT',
        body: JSON.stringify({ name: editName.trim(), is_active: member.is_active }),
      });
      editingId = 0;
      editName = '';
      await loadData();
    } catch (e) {
      error = e.message || 'Không thể cập nhật tên.';
    }
  }

  async function deleteMember(member) {
    if (!confirm(`Xóa ${member.name}? Lịch sử thanh toán sẽ được giữ lại.`)) return;
    try {
      await mealFetch(`/api/meals/members/${member.id}`, { method: 'DELETE' });
      await loadData();
    } catch (e) {
      error = 'Không thể xóa thành viên.';
    }
  }

  async function toggleParticipation(memberId, mealType) {
    const currentBreakfast = isParticipating(memberId, 'breakfast');
    const currentLunch = isParticipating(memberId, 'lunch');

    const newBreakfast = mealType === 'breakfast' ? !currentBreakfast : currentBreakfast;
    const newLunch = mealType === 'lunch' ? !currentLunch : currentLunch;

    try {
      await mealFetch('/api/meals/participations', {
        method: 'PUT',
        body: JSON.stringify({ member_id: memberId, breakfast: newBreakfast, lunch: newLunch }),
      });
      await loadData();
    } catch (e) {
      error = 'Không thể cập nhật.';
    }
  }
</script>

{#if loading}
  <div class="mm-loading">Đang tải...</div>
{:else}
  <div class="member-manager">
    <h2>Quản lý thành viên</h2>

    {#if error}
      <p class="mm-error" role="alert">{error}</p>
    {/if}

    <!-- Add member form -->
    <form class="add-form" onsubmit={(e) => { e.preventDefault(); addMember(); }}>
      <input
        type="text"
        bind:value={newName}
        placeholder="Tên thành viên mới"
        maxlength="100"
        aria-label="Tên thành viên mới"
      />
      <button type="submit" disabled={!newName.trim()}>+ Thêm</button>
    </form>

    <!-- Members list -->
    <div class="members-list">
      <div class="members-header">
        <span>Tên</span>
        <span>Sáng</span>
        <span>Trưa</span>
        <span>Active</span>
        <span></span>
      </div>
      {#each members as member}
        <div class="member-row" class:inactive={!member.is_active}>
          <span class="member-name">
            {#if editingId === member.id}
              <input
                type="text"
                bind:value={editName}
                onkeydown={(e) => { if (e.key === 'Enter') saveName(member); if (e.key === 'Escape') { editingId = 0; } }}
                class="edit-input"
              />
              <button class="btn-sm" onclick={() => saveName(member)}>💾</button>
              <button class="btn-sm" onclick={() => { editingId = 0; }}>✕</button>
            {:else}
              <span ondblclick={() => { editingId = member.id; editName = member.name; }}>{member.name}</span>
              <button class="btn-sm btn-edit" onclick={() => { editingId = member.id; editName = member.name; }}>✏️</button>
            {/if}
          </span>
          <span>
            <input
              type="checkbox"
              checked={isParticipating(member.id, 'breakfast')}
              onchange={() => toggleParticipation(member.id, 'breakfast')}
              aria-label="{member.name} tham gia bữa sáng"
            />
          </span>
          <span>
            <input
              type="checkbox"
              checked={isParticipating(member.id, 'lunch')}
              onchange={() => toggleParticipation(member.id, 'lunch')}
              aria-label="{member.name} tham gia bữa trưa"
            />
          </span>
          <span>
            <input
              type="checkbox"
              checked={member.is_active}
              onchange={() => toggleActive(member)}
              aria-label="{member.name} active"
            />
          </span>
          <span>
            <button class="btn-delete" onclick={() => deleteMember(member)} aria-label="Xóa {member.name}">🗑️</button>
          </span>
        </div>
      {/each}
    </div>

    <p class="mm-hint">💡 Double-click tên để sửa. Bỏ tick "Active" để tạm loại khỏi lịch.</p>
  </div>
{/if}

<style>
  .mm-loading {
    text-align: center;
    padding: 2rem;
    color: var(--text-muted, #718096);
  }

  .member-manager {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  h2 {
    font-size: 1.25rem;
    margin: 0;
    color: var(--text-color, #1a202c);
  }

  .mm-error {
    color: #ef4444;
    font-size: 0.875rem;
    margin: 0;
  }

  .add-form {
    display: flex;
    gap: 0.5rem;
  }

  .add-form input {
    flex: 1;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border-color, #e2e8f0);
    border-radius: 0.375rem;
    font-size: 0.875rem;
    background: var(--input-bg, #ffffff);
    color: var(--text-color, #1a202c);
  }

  .add-form button {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 0.375rem;
    background: var(--primary-color, #3b82f6);
    color: #ffffff;
    font-size: 0.875rem;
    cursor: pointer;
    white-space: nowrap;
  }

  .add-form button:disabled { opacity: 0.5; cursor: not-allowed; }

  .members-list {
    border: 1px solid var(--border-color, #e2e8f0);
    border-radius: 0.5rem;
    overflow: hidden;
  }

  .members-header, .member-row {
    display: grid;
    grid-template-columns: 1fr 60px 60px 60px 40px;
    align-items: center;
    padding: 0.5rem 0.75rem;
    gap: 0.5rem;
  }

  .members-header {
    background: var(--header-bg, #f7fafc);
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--text-muted, #718096);
    text-transform: uppercase;
  }

  .member-row {
    border-top: 1px solid var(--border-color, #e2e8f0);
    font-size: 0.875rem;
  }

  .member-row.inactive {
    opacity: 0.5;
  }

  .member-name {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .edit-input {
    padding: 0.25rem 0.5rem;
    border: 1px solid var(--primary-color, #3b82f6);
    border-radius: 0.25rem;
    font-size: 0.875rem;
    width: 120px;
  }

  .btn-sm {
    padding: 0.125rem 0.25rem;
    border: none;
    background: transparent;
    cursor: pointer;
    font-size: 0.75rem;
  }

  .btn-edit {
    opacity: 0;
    transition: opacity 0.2s;
  }

  .member-name:hover .btn-edit {
    opacity: 1;
  }

  .btn-delete {
    padding: 0.125rem 0.25rem;
    border: none;
    background: transparent;
    cursor: pointer;
    opacity: 0.5;
    transition: opacity 0.2s;
  }

  .btn-delete:hover { opacity: 1; }

  .mm-hint {
    font-size: 0.75rem;
    color: var(--text-muted, #718096);
    margin: 0;
  }

  :global(.dark) .members-header {
    background: var(--header-bg, #1e293b);
  }

  :global(.dark) .members-list {
    border-color: var(--border-color, #334155);
  }

  :global(.dark) .member-row {
    border-color: var(--border-color, #334155);
  }

  :global(.dark) h2 {
    color: var(--text-color, #f1f5f9);
  }

  :global(.dark) .add-form input,
  :global(.dark) .edit-input {
    background: var(--input-bg, #0f172a);
    border-color: var(--border-color, #334155);
    color: var(--text-color, #f1f5f9);
  }
</style>
