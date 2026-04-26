import DealsListPage from "@/pages/deals/DealsListPage.tsx";
import type { DealsListMode } from "@/pages/deals/dealsListModes.ts";

interface DealsStatusBoardPageProps {
  mode: Extract<DealsListMode, "active" | "history">;
}

function DealsStatusBoardPage({ mode }: DealsStatusBoardPageProps) {
  return <DealsListPage mode={mode} />;
}

export default DealsStatusBoardPage;
