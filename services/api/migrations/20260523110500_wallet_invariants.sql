-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION prevent_negative_point_balance()
RETURNS TRIGGER AS $$
DECLARE
  current_balance INT;
BEGIN
  IF NEW.delta_points >= 0 THEN
    RETURN NEW;
  END IF;

  -- Serialize debits per user so concurrent transactions cannot both pass the balance check.
  PERFORM 1 FROM users WHERE id = NEW.user_id FOR UPDATE;

  SELECT COALESCE(SUM(delta_points), 0)::int
    INTO current_balance
    FROM ledger_entries
   WHERE user_id = NEW.user_id;

  IF current_balance + NEW.delta_points < 0 THEN
    RAISE EXCEPTION 'insufficient_points';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS ledger_entries_prevent_negative_balance ON ledger_entries;
CREATE TRIGGER ledger_entries_prevent_negative_balance
BEFORE INSERT ON ledger_entries
FOR EACH ROW
EXECUTE FUNCTION prevent_negative_point_balance();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS ledger_entries_prevent_negative_balance ON ledger_entries;
DROP FUNCTION IF EXISTS prevent_negative_point_balance();
-- +goose StatementEnd
