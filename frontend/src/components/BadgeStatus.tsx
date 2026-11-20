import type { Status } from "../api/types";
import { NOME_STATUS } from "../utils/rotulos";

export function BadgeStatus({ status }: { status: Status }) {
  return <span className={`badge badge-${status}`}>{NOME_STATUS[status]}</span>;
}
