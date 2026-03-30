-- Trigger for bounty fulfillment
DROP TRIGGER IF EXISTS trg_bounty_fulfillment ON bounty_offer;

CREATE TRIGGER trg_bounty_fulfillment
AFTER UPDATE ON bounty_offer
FOR EACH ROW
EXECUTE FUNCTION fn_fulfill_bounty_offer();
