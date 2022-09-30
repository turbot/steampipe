import Icon from "../../../Icon";
import Modal from "../../../Modal";
import React, { useState } from "react";
import useCheckGroupingConfig from "../../../../hooks/useCheckGroupingConfig";
import { CheckDisplayGroup } from "../common";
import NeutralButton from "../../../forms/NeutralButton";

interface CheckGroupingEditorProps {
  config: CheckDisplayGroup[];
  setShowEditor: (show: boolean) => void;
}

interface CheckGroupingTitleLabelProps {
  item: CheckDisplayGroup;
}

const CheckGroupingEditor = ({
  config,
  setShowEditor,
}: CheckGroupingEditorProps) => {
  return (
    <Modal
      actions={[
        <NeutralButton onClick={() => setShowEditor(false)}>
          <>Cancel</>
        </NeutralButton>,
      ]}
      icon={<Icon className="h-6 w-6" icon="pencil-square" />}
      onClose={async () => {
        setShowEditor(false);
      }}
      title="Edit benchmark grouping"
    >
      <div>Foo</div>
    </Modal>
  );
};

const CheckGroupingTitleLabel = ({ item }: CheckGroupingTitleLabelProps) => {
  switch (item.type) {
    case "dimension":
    case "tag":
      return (
        <div>
          <span className="capitalize">{item.type}</span>
          <span>=</span>
          <span>{item.value}</span>
        </div>
      );
    default:
      return (
        <div>
          <span className="capitalize">{item.type}</span>
        </div>
      );
  }
};

const CheckGroupingConfig = ({}) => {
  const [showEditor, setShowEditor] = useState(false);
  const groupingConfig = useCheckGroupingConfig();
  return (
    <>
      <div className="flex items-center space-x-2 shrink-0">
        <span className="font-medium">Grouping:</span>
        {groupingConfig
          .map<React.ReactNode>((item) => (
            <CheckGroupingTitleLabel
              key={`${item.type}${item.value ? `-${item.value}` : ""}`}
              item={item}
            />
          ))
          .reduce((prev, curr, idx) => [
            prev,
            <Icon key={idx} className="h-4 w-4" icon="arrow-long-right" />,
            curr,
          ])}
        <Icon
          className="h-4 w-4 text-foreground-light hover:text-foreground"
          icon="pencil-square"
          onClick={() => setShowEditor(true)}
          title="Edit grouping"
        />
      </div>
      {showEditor && (
        <CheckGroupingEditor
          config={groupingConfig}
          setShowEditor={setShowEditor}
        />
      )}
    </>
  );
};

export default CheckGroupingConfig;
