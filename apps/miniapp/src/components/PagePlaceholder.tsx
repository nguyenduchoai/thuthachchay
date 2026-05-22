import { Link } from 'react-router-dom';

// Component placeholder dùng tạm cho các page chưa có UI thật.
// Sẽ thay dần bằng implementation chính thức ở các milestone sau.
export function PagePlaceholder({
  title,
  hint,
  next,
}: {
  title: string;
  hint?: string;
  next?: { to: string; label: string };
}) {
  return (
    <section className="page-placeholder">
      <h1>{title}</h1>
      {hint && <p className="page-placeholder__hint">{hint}</p>}
      {next && (
        <Link to={next.to} className="btn btn--primary">
          {next.label}
        </Link>
      )}
    </section>
  );
}
