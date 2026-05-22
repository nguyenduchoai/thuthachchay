import { useParams } from 'react-router-dom';
import { PagePlaceholder } from '@/components/PagePlaceholder';

export default function Checkout() {
  const { tx } = useParams<{ tx: string }>();
  return (
    <PagePlaceholder
      title="Checkout…"
      hint={`Poll GET /v1/transactions/${tx ?? ':tx'} mỗi 1s đến status=complete`}
    />
  );
}
