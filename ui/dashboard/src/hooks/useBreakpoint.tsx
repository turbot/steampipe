import React from "react";
import useWindowSize from "./useWindowSize";
import { useCallback, useEffect, useState } from "react";

interface IBreakpointContext {
  currentBreakpoint: string | null;
  maxBreakpoint(breakpointAndDown: string): boolean;
  minBreakpoint(breakpointAndUp: string): boolean;
  width: number;
}

const BreakpointContext = React.createContext<IBreakpointContext | null>(null);

const smBreakpoint = 640;
const mdBreakpoint = 768;
const lgBreakpoint = 1024;
const xlBreakpoint = 1280;
const xxlBreakpoint = 1536;

const getBreakpoint = (width: number) => {
  if (width === 0) return null;
  if (width < smBreakpoint) return "xs";
  if (width < mdBreakpoint) return "sm";
  if (width < lgBreakpoint) return "md";
  if (width < xlBreakpoint) return "lg";
  if (width < xxlBreakpoint) return "xl";
  if (width >= xxlBreakpoint) return "2xl";
  return null;
};

const checkMaxBreakpoint = (
  currentBreakpoint: string | null,
  breakpointAndDown: string
): boolean => {
  // If we have no current breakpoint, return false
  if (!currentBreakpoint) return false;
  // We always display xxl
  if (breakpointAndDown === "2xl") return true;
  // If xl and down, then check if current breakpoint is less than xl
  if (breakpointAndDown === "xl") {
    return currentBreakpoint !== "2xl";
  }
  // If lg and down, then check if current breakpoint is less than xl
  if (breakpointAndDown === "lg") {
    return currentBreakpoint !== "xl";
  }
  // If md and down, then check if current breakpoint is less than lg
  if (breakpointAndDown === "md") {
    return !(currentBreakpoint === "lg" || currentBreakpoint === "xl");
  }
  // If sm and down, then check if current breakpoint is less than md
  if (breakpointAndDown === "sm") {
    return !(
      currentBreakpoint === "md" ||
      currentBreakpoint === "lg" ||
      currentBreakpoint === "xl"
    );
  }
  // If xs and down, then check if current breakpoint is less than sm
  if (breakpointAndDown === "xs") {
    return !(
      currentBreakpoint === "sm" ||
      currentBreakpoint === "md" ||
      currentBreakpoint === "lg" ||
      currentBreakpoint === "xl"
    );
  }
  // Else it's an unknown breakpoint and down, so return false
  return false;
};

const checkMinBreakpoint = (
  currentBreakpoint: string | null,
  breakpointAndUp: string
): boolean => {
  // If we have no current breakpoint, return false
  if (!currentBreakpoint) return false;
  // We always display xs
  if (breakpointAndUp === "xs") return true;
  // If sm and up, then check if current breakpoint is less than sm
  if (breakpointAndUp === "sm") {
    return currentBreakpoint !== "xs";
  }
  // If md and up, then check if current breakpoint is less than md
  if (breakpointAndUp === "md") {
    return !(currentBreakpoint === "xs" || currentBreakpoint === "sm");
  }
  // If lg and up, then check if current breakpoint is less than lg
  if (breakpointAndUp === "lg") {
    return !(
      currentBreakpoint === "xs" ||
      currentBreakpoint === "sm" ||
      currentBreakpoint === "md"
    );
  }
  // If xl and up, then check if current breakpoint is less than xl
  if (breakpointAndUp === "xl") {
    return !(
      currentBreakpoint === "xs" ||
      currentBreakpoint === "sm" ||
      currentBreakpoint === "md" ||
      currentBreakpoint === "lg"
    );
  }
  // If xl and up, then check if current breakpoint is less than xl
  if (breakpointAndUp === "2xl") {
    return !(
      currentBreakpoint === "xs" ||
      currentBreakpoint === "sm" ||
      currentBreakpoint === "md" ||
      currentBreakpoint === "lg" ||
      currentBreakpoint === "xl"
    );
  }
  // Else it's an unknown breakpoint and up, so return false
  return false;
};

const BreakpointProvider = ({ children }: { children: React.ReactNode }) => {
  const [width] = useWindowSize();
  const [currentBreakpoint, setCurrentBreakpoint] = useState(
    getBreakpoint(width)
  );
  useEffect(() => {
    setCurrentBreakpoint(getBreakpoint(width));
  }, [width]);
  const maxBreakpoint = useCallback(
    (breakpointAndDown) =>
      checkMaxBreakpoint(currentBreakpoint, breakpointAndDown),
    [currentBreakpoint]
  );
  const minBreakpoint = useCallback(
    (breakpointAndUp) => checkMinBreakpoint(currentBreakpoint, breakpointAndUp),
    [currentBreakpoint]
  );
  return (
    <BreakpointContext.Provider
      value={{
        currentBreakpoint,
        maxBreakpoint,
        minBreakpoint,
        width,
      }}
    >
      {children}
    </BreakpointContext.Provider>
  );
};

const useBreakpoint = () => {
  const context = React.useContext(BreakpointContext);
  if (context === undefined) {
    throw new Error("useBreakpoint must be used within a BreakpointContext");
  }
  return context as IBreakpointContext;
};

export { BreakpointProvider, useBreakpoint };
