import Grid from "../Grid";
import NeutralButton from "../../../forms/NeutralButton";
import PanelDetailData from "./PanelDetailData";
import PanelDetailDataDownloadButton from "./PanelDetailDataDownloadButton";
import PanelDetailDefinition from "./PanelDetailDefinition";
import PanelDetailPreview from "./PanelDetailPreview";
import PanelDetailQuery from "./PanelDetailQuery";
import { classNames } from "../../../../utils/styles";
import { PanelDefinition } from "../../../../types";
import { useDashboard } from "../../../../hooks/useDashboard";
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
  QUERY: {
    name: "query",
    label: "Query",
    Component: PanelDetailQuery,
  },
  DEFINITION: {
    name: "definition",
    label: "Definition",
    Component: PanelDetailDefinition,
  },
  DATA: {
    name: "data",
    label: "Data",
    Component: PanelDetailData,
  },
};

const PanelDetail = ({ definition }: PanelDetailProps) => {
  const [selectedTab, setSelectedTab] = useState(Tabs.PREVIEW);
  const {
    breakpointContext: { minBreakpoint },
    closePanelDetail,
  } = useDashboard();
  const isTablet = minBreakpoint("md");

  const availableTabs = useMemo(() => {
    const tabs = [
      {
        ...Tabs.PREVIEW,
        selected: selectedTab.name === Tabs.PREVIEW.name,
      },
    ];
    if (definition.sql) {
      tabs.push({
        ...Tabs.QUERY,
        selected: selectedTab.name === Tabs.QUERY.name,
      });
    }
    if (definition.source_definition) {
      tabs.push({
        ...Tabs.DEFINITION,
        selected: selectedTab.name === Tabs.DEFINITION.name,
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

  return (
    <div className="h-full overflow-y-auto p-4">
      <Grid name={definition.name}>
        <div className="col-span-6">
          <h2 className="break-all">{definition.title || "Panel Detail"}</h2>
        </div>
        <div className="col-span-6 space-x-2 text-right">
          <PanelDetailDataDownloadButton
            panelDefinition={definition}
            size={isTablet ? "md" : "sm"}
          />
          <NeutralButton
            onClick={closePanelDetail}
            size={isTablet ? "md" : "sm"}
          >
            <>
              Close<span className="ml-2 font-light text-xxs">ESC</span>
            </>
          </NeutralButton>
        </div>
        <div className="col-span-12 sm:hidden ">
          <label htmlFor="tabs" className="sr-only">
            Select a tab
          </label>
          {/* Use an "onChange" listener to redirect the user to the selected tab URL. */}
          <select
            id="tabs"
            name="tabs"
            className="mt-2 block w-full pl-3 pr-10 py-2 bg-dashboard print:bg-white text-foreground border-black-scale-3 focus:outline-none focus:ring-purple-500 focus:border-purple-500 sm:text-sm rounded-md"
            defaultValue={selectedTab.name}
            onChange={(e) =>
              setSelectedTab(
                availableTabs.find((tab) => tab.name === e.target.value) ||
                  availableTabs[0]
              )
            }
          >
            {availableTabs.map((tab) => (
              <option key={tab.name} value={tab.name}>
                {tab.label}
              </option>
            ))}
          </select>
        </div>
        <div className="col-span-12 hidden sm:block">
          <div className="border-b border-black-scale-3">
            <nav className="-mb-px flex space-x-6" aria-label="Tabs">
              {availableTabs.map((tab) => (
                <span
                  key={tab.name}
                  className={classNames(
                    tab.selected
                      ? "border-black-scale-4 text-foreground cursor-pointer"
                      : "border-transparent text-foreground-lighter hover:text-foreground cursor-pointer",
                    "whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm"
                  )}
                  onClick={() => setSelectedTab(tab)}
                >
                  {tab.label}
                </span>
              ))}
            </nav>
          </div>
        </div>

        <div className="col-span-12 mt-4">
          {<selectedTab.Component definition={definition} />}
        </div>
      </Grid>
    </div>
  );
};

export default PanelDetail;
