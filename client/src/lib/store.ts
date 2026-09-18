import { configureStore } from "@reduxjs/toolkit";
import appReducer from "./slices/app";
import authReducer from "./slices/auth";

export const makeStore = () =>
  configureStore({
    reducer: {
      app: appReducer,
      auth: authReducer,
    },
  });

export type AppStore = ReturnType<typeof makeStore>;
export type RootState = ReturnType<AppStore["getState"]>;
export type AppDispatch = AppStore["dispatch"];
