import { useParams } from 'react-router-dom';
import { PagePlaceholder } from '@/components/PagePlaceholder';

export default function ChallengeDetail() {
  const { id } = useParams<{ id: string }>();
  return (
    <PagePlaceholder
      title="Challenge detail"
      hint={`GET /v1/challenges/${id ?? ':id'} → stats grid + sticky CTA Join`}
    />
  );
}
