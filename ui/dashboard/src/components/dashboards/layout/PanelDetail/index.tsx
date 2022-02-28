import LayoutPanel from "../common/LayoutPanel";
import NeutralButton from "../../../forms/NeutralButton";
import PanelDetailData from "./PanelDetailData";
import PanelDetailDefinition from "./PanelDetailDefinition";
import PanelDetailPreview from "./PanelDetailPreview";
import PanelDetailQuery from "./PanelDetailQuery";
import { classNames } from "../../../../utils/styles";
import { PanelDefinition, useDashboard } from "../../../../hooks/useDashboard";
import { useMemo, useState } from "react";

export type PanelDetailProps = {
  definition: PanelDefinition;
};

const Tabs = {
  PREVIEW: {
    name: "preview",
    label: "Preview",
    Component: PanelDetailPreview,
  },
  DEFINITION: {
    name: "definition",
    label: "Definition",
    Component: PanelDetailDefinition,
  },
  QUERY: {
    name: "query",
    label: "Query",
    Component: PanelDetailQuery,
  },
  DATA: {
    name: "data",
    label: "Data",
    Component: PanelDetailData,
  },
};

const PanelDetail = ({ definition }: PanelDetailProps) => {
  const [selectedTab, setSelectedTab] = useState(Tabs.PREVIEW);
  const availableTabs = useMemo(() => {
    const tabs = [
      {
        ...Tabs.PREVIEW,
        selected: selectedTab.name === Tabs.PREVIEW.name,
      },
    ];
    if (definition.source_definition) {
      tabs.push({
        ...Tabs.DEFINITION,
        selected: selectedTab.name === Tabs.DEFINITION.name,
      });
    }
    if (definition.sql) {
      tabs.push({
        ...Tabs.QUERY,
        selected: selectedTab.name === Tabs.QUERY.name,
      });
    }
    if (definition.data) {
      tabs.push({
        ...Tabs.DATA,
        selected: selectedTab.name === Tabs.DATA.name,
      });
    }
    return tabs;
  }, [definition, selectedTab]);
  const { closePanelDetail } = useDashboard();
  return (
    <LayoutPanel definition={definition} withPadding={true}>
      <div className="col-span-11">
        <h2 className="text-2xl font-medium break-all">Panel Detail</h2>
      </div>
      <div className="col-span-1 text-right">
        <NeutralButton onClick={closePanelDetail}>
          <>
            Close<span className="ml-2 font-light text-xxs">ESC</span>
          </>
        </NeutralButton>
      </div>
      <div className="col-span-6 sm:hidden">
        <label htmlFor="tabs" className="sr-only">
          Select a tab
        </label>
        {/* Use an "onChange" listener to redirect the user to the selected tab URL. */}
        <select
          id="tabs"
          name="tabs"
          className="mt-4 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-purple-500 focus:border-purple-500 sm:text-sm rounded-md"
          defaultValue={selectedTab.name}
        >
          {availableTabs.map((tab) => (
            <option key={tab.name}>{tab.label}</option>
          ))}
        </select>
      </div>
      <div className="col-span-12 hidden sm:block">
        <div className="border-b border-gray-200">
          <nav className="mt-2 -mb-px flex space-x-8" aria-label="Tabs">
            {availableTabs.map((tab) => (
              <span
                key={tab.name}
                className={classNames(
                  tab.selected
                    ? "border-purple-500 text-purple-600"
                    : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-200",
                  "whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm"
                )}
                onClick={() => setSelectedTab(tab)}
              >
                {tab.label}
              </span>
            ))}
          </nav>
        </div>
      </div>

      <>{<selectedTab.Component definition={definition} />}</>

      {/*<div className="col-span-12 grid grid-cols-12">*/}
      {/*  <div className="col-span-12 md:col-span-6 lg:col-span-4">*/}
      {/*    {definition.sql && <PanelQuery query={definition.sql} />}*/}
      {/*  </div>*/}
      {/*  <div className="col-span-12 md:col-span-6 lg:col-span-8">*/}
      {/*    {definition.data && (*/}
      {/*      <Table*/}
      {/*        name={`${definition}.table.detail`}*/}
      {/*        node_type="table"*/}
      {/*        data={definition.data}*/}
      {/*      />*/}
      {/*    )}*/}
      {/*  </div>*/}
      {/*</div>*/}
    </LayoutPanel>
  );
};

export default PanelDetail;
