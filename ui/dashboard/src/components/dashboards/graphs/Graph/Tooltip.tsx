import {
  cloneElement,
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";
import { ThemeProvider, ThemeWrapper } from "../../../../hooks/useTheme";
import { usePopper } from "react-popper";
import { v4 as uuidv4 } from "uuid";
import { noop } from "../../../../utils/func";
import { CheckGroupingContext } from "../../../../hooks/useCheckGrouping";

interface TooltipProps {
  children: JSX.Element;
  overlay?: JSX.Element | JSX.Element[];
  show?: boolean;
  title: string;
}

// Start: adapted from https://github.com/streamich/react-use
export function on<T extends Window | Document | HTMLElement | EventTarget>(
  obj: T | null,
  ...args: Parameters<T["addEventListener"]> | [string, Function | null, ...any]
): void {
  if (obj && obj.addEventListener) {
    obj.addEventListener(
      ...(args as Parameters<HTMLElement["addEventListener"]>)
    );
  }
}

export function off<T extends Window | Document | HTMLElement | EventTarget>(
  obj: T | null,
  ...args:
    | Parameters<T["removeEventListener"]>
    | [string, Function | null, ...any]
): void {
  if (obj && obj.removeEventListener) {
    obj.removeEventListener(
      ...(args as Parameters<HTMLElement["removeEventListener"]>)
    );
  }
}

const defaultEvents = ["mousedown", "touchstart"];

const useClickAway = <E extends Event = Event>(
  ref: Element | null,
  onClickAway: (event: E) => void,
  events: string[] = defaultEvents
) => {
  const savedCallback = useRef(onClickAway);
  useEffect(() => {
    savedCallback.current = onClickAway;
  }, [onClickAway]);
  useEffect(() => {
    if (!ref) {
      return;
    }
    const handler = (event) => {
      ref && !ref.contains(event.target) && savedCallback.current(event);
    };
    for (const eventName of events) {
      on(document, eventName, handler);
    }
    return () => {
      for (const eventName of events) {
        off(document, eventName, handler);
      }
    };
  }, [events, ref]);
};
// End: adapted from https://github.com/streamich/react-use

interface ITooltipsContext {
  closeTooltips: (id?: string) => void;
  retainTooltipId: string | null;
  shouldCloseTooltips: boolean;
}

const TooltipsContext = createContext<ITooltipsContext | null>({
  closeTooltips: noop,
  retainTooltipId: null,
  shouldCloseTooltips: false,
});

const TooltipsProvider = ({ children }) => {
  const [shouldCloseTooltips, setShouldCloseTooltips] = useState(false);
  const [retainTooltipId, setRetainTooltipId] = useState<string | null>(null);

  const closeTooltips = (id?: string) => {
    if (id) {
      setRetainTooltipId(id);
    } else {
      setRetainTooltipId(null);
    }
    setShouldCloseTooltips(true);
  };

  return (
    <TooltipsContext.Provider
      value={{ closeTooltips, retainTooltipId, shouldCloseTooltips }}
    >
      {children}
    </TooltipsContext.Provider>
  );
};

const useTooltips = () => {
  const context = useContext(TooltipsContext);
  if (context === undefined) {
    throw new Error("useTooltips must be used within a TooltipsContext");
  }
  return context as ITooltipsContext;
};

const Tooltip = ({ children, overlay, show = false, title }: TooltipProps) => {
  const [id] = useState(uuidv4());
  const [showOverlay, setShowOverlay] = useState(false);
  const [referenceElement, setReferenceElement] = useState(null);
  const [popperElement, setPopperElement] = useState(null);
  const [arrowElement, setArrowElement] = useState(null);
  const { styles, attributes } = usePopper(referenceElement, popperElement, {
    modifiers: [{ name: "arrow", options: { element: arrowElement } }],
  });
  const { closeTooltips, retainTooltipId, shouldCloseTooltips } = useTooltips();

  const trigger = cloneElement(children, {
    ref: setReferenceElement,
    onMouseEnter: overlay
      ? () => {
          closeTooltips(id);
          setShowOverlay(true);
        }
      : undefined,
    onTouchStart: overlay
      ? () => {
          closeTooltips(id);
          setShowOverlay(true);
        }
      : undefined,
    // onMouseLeave: overlay ? () => setShowOverlay(false) : undefined,
  });

  useEffect(() => {
    if (!shouldCloseTooltips || retainTooltipId === id) {
      return;
    }
    setShowOverlay(false);
  }, [id, retainTooltipId, shouldCloseTooltips]);

  // @ts-ignore
  useClickAway(popperElement, () => setShowOverlay(false));

  return (
    <>
      {trigger}
      {(show || showOverlay) &&
        createPortal(
          <ThemeProvider>
            <ThemeWrapper>
              <div
                // @ts-ignore
                ref={setPopperElement}
                style={styles.popper}
                {...attributes.popper}
                className="z-50 bg-dashboard p-3 border border-divide rounded-md text-sm flex flex-col space-y-2 bg-dashboard-panel"
              >
                <Title title={title} />
                {overlay}
                {/*@ts-ignore*/}
                <div ref={setArrowElement} style={styles.arrow} />
              </div>
            </ThemeWrapper>
          </ThemeProvider>,
          document.body
        )}
    </>
  );
};

const Title = ({ title }) => {
  return <strong className="block break-all">{title}</strong>;
};

export default Tooltip;

export { TooltipsProvider, useTooltips };
