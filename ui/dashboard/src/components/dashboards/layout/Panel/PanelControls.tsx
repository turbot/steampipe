import Icon from "../../../Icon";
import { createPortal } from "react-dom";
import { ThemeProvider, ThemeWrapper } from "../../../../hooks/useTheme";
import { usePopper } from "react-popper";
import { useState } from "react";

const PanelControl = ({ action, icon, title }) => {
  return (
    <div
      className="p-1 cursor-pointer bg-dashboard-panel text-foreground"
      onClick={async (e) => await action(e)}
      title={title}
    >
      <Icon className="w-5 h-5" icon={icon} />
    </div>
  );
};

const PanelControls = ({ controls, referenceElement }) => {
  const [popperElement, setPopperElement] = useState(null);
  const { styles, attributes } = usePopper(referenceElement, popperElement, {
    placement: "top-end",
  });
  return createPortal(
    <ThemeProvider>
      <ThemeWrapper>
        <div
          // @ts-ignore
          ref={setPopperElement}
          // className={classNames("z-50")}
          style={{ ...styles.popper }}
          {...attributes.popper}
        >
          <div className="flex space-x-px border border-black-scale-3">
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
    document.body
  );
};

export default PanelControls;
