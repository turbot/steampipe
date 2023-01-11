import Icon from "../../../Icon";
import { createPortal } from "react-dom";
import { ThemeProvider, ThemeWrapper } from "../../../../hooks/useTheme";
import { useMemo, useState } from "react";
import { usePopper } from "react-popper";

export type IPanelControl = {
  action: (e: any) => Promise<void>;
  icon: string;
  title: string;
};

const PanelControl = ({ action, icon, title }: IPanelControl) => {
  return (
    <div
      className="p-1.5 cursor-pointer bg-dashboard-panel text-foreground first:rounded-tl-[4px] first:rounded-bl-[4px] last:rounded-tr-[4px] last:rounded-br-[4px] hover:bg-dashboard"
      onClick={async (e) => await action(e)}
      title={title}
    >
      <Icon className="w-4.5 h-4.5" icon={icon} />
    </div>
  );
};

const PanelControls = ({ controls, referenceElement }) => {
  const [popperElement, setPopperElement] = useState(null);
  // Need to define memoized / stable modifiers else the usePopper hook will infinitely re-render
  const noFlip = useMemo(() => ({ name: "flip", enabled: false }), []);
  const offset = useMemo(
    () => ({
      name: "offset",
      options: {
        // For some reason the height of the popper is not correct unless scrollbars are visible.
        // I've sunk too much time trying to find the root cause, but luckily I only
        // need to modify this along a fixed offset, so can hard-code this for now.
        offset: [-14.125, -14.125],
        // offset: ({ popper }) => {
        // const offset = -popper.height / 2;
        // return [offset, offset];
        // },
      },
    }),
    []
  );
  const { styles, attributes } = usePopper(referenceElement, popperElement, {
    modifiers: [noFlip, offset],
    placement: "top-end",
  });

  return createPortal(
    <ThemeProvider>
      <ThemeWrapper>
        <div
          // @ts-ignore
          ref={setPopperElement}
          style={{ ...styles.popper }}
          {...attributes.popper}
        >
          <div className="flex border border-black-scale-3 rounded-md">
            {controls.map((control, idx) => (
              <PanelControl
                key={idx}
                action={control.action}
                icon={control.icon}
                title={control.title}
              />
            ))}
          </div>
        </div>
      </ThemeWrapper>
    </ThemeProvider>,
    // @ts-ignore as this element definitely exists
    document.getElementById("portals")
  );
};

export default PanelControls;
