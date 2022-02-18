import { BreakpointProvider } from "../src/hooks/useBreakpoint";
import { ThemeWrapper } from "../src/hooks/useTheme";

const StoryWrapper = ({ children }) => {
  return (
    <BreakpointProvider>
      <ThemeWrapper>{children}</ThemeWrapper>
    </BreakpointProvider>
  );
};

export default StoryWrapper;
