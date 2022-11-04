import Icon from "../../../Icon";

const PanelControl = ({ action, icon, title }) => {
  return (
    <div
      className="p-1 cursor-pointer bg-black-scale-2 text-foreground first:rounded-tl-[4px] first:rounded-bl-[4px] last:rounded-tr-[4px] last:rounded-br-[4px]"
      onClick={async (e) => await action(e)}
      title={title}
    >
      <Icon className="w-5 h-5" icon={icon} />
    </div>
  );
};

const PanelControls = ({ controls }) => (
  <div className="flex space-x-px">
    {controls.map((control, idx) => (
      <PanelControl
        key={idx}
        action={control.action}
        icon={control.icon}
        title={control.title}
      />
    ))}
  </div>
);

export default PanelControls;
