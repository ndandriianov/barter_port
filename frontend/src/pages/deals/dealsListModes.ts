import type { DealStatus } from "@/features/deals/model/types.ts";

export type DealsListMode = "active" | "joinable" | "history";

export interface DealsListSelection {
  key: string;
  title: string;
  description: string;
  statuses: DealStatus[];
}

export interface DealsListModeConfig {
  eyebrow: string;
  title: string;
  description: string;
  defaultStatuses: DealStatus[];
  query: {
    myOnly: boolean;
    openOnly?: boolean;
    excludeCurrentUser?: boolean;
  };
  selections?: DealsListSelection[];
}

const activeStatuses: DealStatus[] = ["LookingForParticipants", "Discussion", "Confirmed"];
const joinableStatuses: DealStatus[] = ["LookingForParticipants"];
const historyStatuses: DealStatus[] = ["Completed", "Cancelled", "Failed"];

export const dealsListModeConfig: Record<DealsListMode, DealsListModeConfig> = {
  active: {
    eyebrow: "Сделки / Активные",
    title: "Активные сделки",
    description: "Только ваши текущие сделки: набор участников, обсуждение и процесс обмена.",
    defaultStatuses: activeStatuses,
    query: {
      myOnly: true,
    },
    selections: [
      {
        key: "LookingForParticipants",
        title: "В поиске участников",
        description: "Ваши сделки, в которые ещё можно добирать участников.",
        statuses: ["LookingForParticipants"],
      },
      {
        key: "Discussion",
        title: "Обсуждение",
        description: "Сделки, которые уже перешли к согласованию условий.",
        statuses: ["Discussion"],
      },
      {
        key: "Confirmed",
        title: "В процессе обмена",
        description: "Сделки, где договорённости подтверждены и идёт обмен.",
        statuses: ["Confirmed"],
      },
    ],
  },
  joinable: {
    eyebrow: "Сделки / Можно присоединиться",
    title: "Можно присоединиться",
    description: "Чужие открытые сделки, в которые можно подать заявку на участие.",
    defaultStatuses: joinableStatuses,
    query: {
      myOnly: false,
      openOnly: true,
      excludeCurrentUser: true,
    },
  },
  history: {
    eyebrow: "Сделки / История",
    title: "История сделок",
    description: "Только ваши завершённые, отменённые и несостоявшиеся сделки.",
    defaultStatuses: historyStatuses,
    query: {
      myOnly: true,
    },
    selections: [
      {
        key: "history",
        title: "Вся история",
        description: "Объединение статусов: завершены, отменены и не состоялись.",
        statuses: historyStatuses,
      },
      {
        key: "Completed",
        title: "Завершены",
        description: "Успешно завершённые сделки.",
        statuses: ["Completed"],
      },
      {
        key: "Cancelled",
        title: "Отменены",
        description: "Сделки, которые были отменены.",
        statuses: ["Cancelled"],
      },
      {
        key: "Failed",
        title: "Не состоялись",
        description: "Сделки, завершившиеся неуспешно.",
        statuses: ["Failed"],
      },
    ],
  },
};
