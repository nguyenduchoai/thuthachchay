import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { createChallenge } from '@/services/endpoints';

const PUBLIC_FEE = 100;

export default function CreateNew() {
  const { t } = useTranslation();
  const nav = useNavigate();
  const [visibility, setVisibility] = useState<'private' | 'public'>('private');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [target, setTarget] = useState(10000);
  const [days, setDays] = useState(30);
  const [entry, setEntry] = useState(500);
  const [start, setStart] = useState(() => new Date().toISOString().slice(0, 10));
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const valid = name.trim().length >= 3 && target > 0 && days > 0;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!valid) return;
    setBusy(true);
    setErr(null);
    try {
      const ch = await createChallenge({
        visibility,
        name: name.trim(),
        description: description.trim() || undefined,
        daily_steps_target: target,
        duration_days: days,
        start_date: new Date(start).toISOString(),
        entry_points: entry,
      });
      nav(`/challenges/${ch.id}`);
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="create-new">
      <header className="header-row">
        <button className="link" onClick={() => nav(-1)}>← {t('common.back')}</button>
        <h1>{t('createNew.title', 'Tạo thử thách')}</h1>
        <span />
      </header>

      <form onSubmit={submit} className="form">
        <fieldset className="visibility">
          <legend>{t('createNew.visibility', 'Loại thử thách')}</legend>
          <label className={`option glass ${visibility === 'private' ? 'is-selected' : ''}`}>
            <input type="radio" name="vis" value="private" checked={visibility === 'private'} onChange={() => setVisibility('private')} />
            <div>
              <strong>🔒 {t('createNew.private', 'Riêng tư')}</strong>
              <small>{t('createNew.privateDesc', 'Chỉ mời · tối đa 20 · miễn phí')}</small>
            </div>
            <span className="tag tag--green">{t('createNew.free', 'Miễn phí')}</span>
          </label>
          <label className={`option glass ${visibility === 'public' ? 'is-selected' : ''}`}>
            <input type="radio" name="vis" value="public" checked={visibility === 'public'} onChange={() => setVisibility('public')} />
            <div>
              <strong>🌐 {t('createNew.public', 'Công khai')}</strong>
              <small>{t('createNew.publicDesc', 'Hiển thị công khai · không giới hạn · host nhận 10% pool')}</small>
            </div>
            <span className="tag">{PUBLIC_FEE}đ</span>
          </label>
        </fieldset>

        <label className="field">
          <span>{t('createNew.name', 'Tên thử thách')}</span>
          <input type="text" value={name} onChange={(e) => setName(e.target.value)} placeholder="VD: Chạy bộ Hè 2026" maxLength={60} required />
        </label>

        <label className="field">
          <span>{t('createNew.description', 'Mô tả (tuỳ chọn)')}</span>
          <textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Mô tả điều đặc biệt..." rows={3} maxLength={500} />
        </label>

        <div className="row-2">
          <label className="field">
            <span>{t('createNew.target', 'Bước/ngày')}</span>
            <input type="number" value={target} onChange={(e) => setTarget(+e.target.value)} min={1000} max={50000} step={500} />
          </label>
          <label className="field">
            <span>{t('createNew.days', 'Số ngày')}</span>
            <input type="number" value={days} onChange={(e) => setDays(+e.target.value)} min={1} max={365} />
          </label>
        </div>

        <div className="row-2">
          <label className="field">
            <span>{t('createNew.entry', 'Phí tham gia (đ)')}</span>
            <input type="number" value={entry} onChange={(e) => setEntry(+e.target.value)} min={0} max={10000} step={100} />
          </label>
          <label className="field">
            <span>{t('createNew.start', 'Ngày bắt đầu')}</span>
            <input type="date" value={start} onChange={(e) => setStart(e.target.value)} />
          </label>
        </div>

        {err && <p className="error">{err}</p>}

        <div className="cta-sticky">
          <small className="muted">
            {visibility === 'public'
              ? `${t('createNew.feePublic', 'Phí tạo')}: ${PUBLIC_FEE}đ`
              : t('createNew.feeFree', 'Miễn phí tạo private')}
          </small>
          <button type="submit" disabled={!valid || busy} className="btn btn--primary btn--full">
            {busy ? t('common.loading') : `🪙 ${t('createNew.create', 'Tạo & tham gia')} · ${visibility === 'public' ? PUBLIC_FEE : entry}đ`}
          </button>
        </div>
      </form>
    </section>
  );
}
