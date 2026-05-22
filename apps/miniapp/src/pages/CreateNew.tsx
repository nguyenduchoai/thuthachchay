import { PagePlaceholder } from '@/components/PagePlaceholder';

export default function CreateNew() {
  return (
    <PagePlaceholder
      title="Create challenge"
      hint="Form: visibility (private/public), name, desc, cover, goal, days. POST /v1/challenges"
    />
  );
}
