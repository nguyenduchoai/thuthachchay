import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { checkHandle, patchMe } from '@/services/endpoints';
import { useUserStore } from '@/state/user';

export default function ProfileEdit() {
  const { t } = useTranslation();
  const nav = useNavigate();
  const user = useUserStore((s) => s.user);
  const setUser = useUserStore((s) => s.set);
  const [handle, setHandle] = useState(user?.handle ?? '');
  const [name, setName] = useState(user?.display_name ?? '');
  const [goal, setGoal] = useState(user?.daily_goal ?? 10000);
  const [available, setAvailable] = useState<boolean | null>(null);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    setHandle(user?.handle ?? '');
    setName(user?.display_name ?? '');
    setGoal(user?.daily_goal ?? 10000);
  }, [user]);

  useEffect(() => {
    if (!handle || handle === user?.handle) {
      setAvailable(null);
      return;
    }
    const tm = setTimeout(() => {
      checkHandle(handle).then((r) => setAvailable(r.available)).catch(() => setAvailable(null));
    }, 350);
    return () => clearTimeout(tm);
  }, [handle, user?.handle]);

  const handleOK = available !== false;
  const canSave =
    handleOK &&
    name.trim().length > 0 &&
    goal >= 1000 &&
    goal <= 50000 &&
    (handle !== user?.handle || name !== user?.display_name || goal !== user?.daily_goal);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setErr(null);
    try {
      const patch: Record<string, unknown> = {};
      if (handle !== user?.handle) patch.handle = handle;
      if (name !== user?.display_name) patch.display_name = name;
      if (goal !== user?.daily_goal) patch.daily_goal = goal;
      const updated = await patchMe(patch);
      setUser(updated);
      nav('/profile');
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="profile-edit">
      <header className="header-row">
        <button className="link" onClick={() => nav(-1)}>← {t('common.back')}</button>
        <h1>{t('profileEdit.title', 'Chỉnh sửa hồ sơ')}</h1>
        <span />
      </header>

      <div className="avatar-edit">
        <div className="avatar avatar--xl">🦘</div>
        <button type="button" className="btn btn--ghost btn--sm">📷 {t('profileEdit.changePhoto', 'Đổi ảnh')}</button>
      </div>

      <form onSubmit={save} className="form">
        <label className="field">
          <span>{t('profileEdit.handle', '@username công khai')}</span>
          <input
            type="text"
            value={handle}
            onChange={(e) => setHandle(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ''))}
            placeholder="nguyenduchoai"
            minLength={3}
            maxLength={20}
            required
          />
          {available === false && <small className="error">{t('profileEdit.taken', 'Username đã có người dùng')}</small>}
          {available === true && <small className="success">✓ {t('profileEdit.available', 'Username khả dụng')}</small>}
        </label>

        <label className="field">
          <span>{t('profileEdit.name', 'Tên hiển thị')}</span>
          <input type="text" value={name} onChange={(e) => setName(e.target.value)} maxLength={60} required />
        </label>

        <label className="field">
          <span>{t('profileEdit.goal', 'Mục tiêu bước/ngày')}</span>
          <input type="number" value={goal} onChange={(e) => setGoal(+e.target.value)} min={1000} max={50000} step={500} />
        </label>

        <label className="field field--locked">
          <span>{t('profileEdit.email', 'Email')}</span>
          <input type="email" value={user?.email ?? ''} disabled />
          <small className="muted">🔒 {t('profileEdit.emailLocked', 'Email lấy từ Zalo, không sửa được.')}</small>
        </label>

        {err && <p className="error">{err}</p>}

        <button type="submit" disabled={!canSave || busy} className="btn btn--primary btn--full">
          {busy ? t('common.loading') : t('profileEdit.save', 'Lưu thay đổi')}
        </button>
      </form>
    </section>
  );
}
