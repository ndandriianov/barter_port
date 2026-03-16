import {type TypedUseSelectorHook, useDispatch, useSelector } from "react-redux";
import type { AppDispatch } from "@/app/store/store";
import type { RootState } from "@/app/store/rootReducer.ts";

export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;