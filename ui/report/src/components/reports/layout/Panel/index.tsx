import Error from "../../Error";
import Placeholder from "../../Placeholder";
import React, { memo, useEffect, useState } from "react";
import useDimensions from "../../../../hooks/useDimensions";
import { BaseChartProps } from "../../charts";
import { classNames } from "../../../../utils/styles";
import { CounterProps } from "../../Counter";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { PanelDefinition, PanelType } from "../../../../hooks/useReport";
import { PanelProvider } from "../../../../hooks/usePanel";
import { TableProps } from "../../Table";

interface PanelProps {
  children: null | JSX.Element | JSX.Element[];
  definition:
    | BaseChartProps
    | CounterProps
    // | ImageProps
    | PanelDefinition
    | TableProps;
  ready?: boolean;
  showExpand?: boolean;
}

// const getReadyStatusForPanelType = (
//   type: PanelType,
//   definition: PanelDefinition
// ): boolean => {
//   switch (type) {
//     case "benchmark":
//     case "control":
//     case "counter":
//     case "error":
//     case "input":
//     case "text":
//       return true;
//     case "chart":
//     case "table":
//       // console.log("Checking chart", {
//       //   currentData: data,
//       //   newDefinition: definition,
//       // });
//       return !!definition.data;
//     case "image":
//       return definition.sql
//         ? !!definition.data
//         : definition.properties
//         ? !!definition.properties.src
//         : false;
//     default:
//       return false;
//   }
// };

const Panel = ({
  children,
  definition,
  showExpand = true,
  ready = true,
}: PanelProps) => {
  const [panelRef, dimensions] = useDimensions();
  // const { sqlDataMap } = useReport();
  // console.log(sqlDataMap);
  // const [data, setData] = useState(
  //   definition.sql ? sqlDataMap[definition.sql] : null
  // );
  // const [ready, setReady] = useState(
  //   getReadyStatusForPanelType(type, definition)
  // );
  //
  // useEffect(() => {
  //   // const newData = definition.sql ? sqlDataMap[definition.sql] : null;
  //   // if (definition.sql) {
  //   //   console.log(newData);
  //   //   setData(newData);
  //   // }
  //   // if (definition.data) {
  //   //   // console.log(
  //   //   //   "Panel definition now has data",
  //   //   //   definition.name,
  //   //   //   definition.data
  //   //   // );
  //   //   setData(definition.data);
  //   // }
  //   setReady(getReadyStatusForPanelType(type, definition));
  // }, [type, definition]);

  // console.log({ name: definition.name, data, ready });

  // const height = useMemo(() => {
  //   // Fail safe
  //   if (!definition) {
  //     return null;
  //   }
  //
  //   // If we haven't got a ref to the panel yet, use auto
  //   if (!panelRef || !panelRef.current) {
  //     return null;
  //   }
  //
  //   // If the panel defines a height, work out a height relative to its width
  //   if (has(definition, "height")) {
  //     // If the panel width and height are the same, make the height equal to the width
  //     if (definition.width === definition.height) {
  //       return dimensions.width;
  //     }
  //
  //     // If the panel width is 1, then we already have the single unit width
  //     if (definition.width === 1) {
  //       return definition.height * dimensions.width;
  //     }
  //
  //     // Work out what a single grid unit equates to
  //     const panelWidthUnits = Math.min(definition.width || 12, 12);
  //     const unitWidth = dimensions.width / panelWidthUnits;
  //
  //     return definition.height * unitWidth;
  //   } // Else use auto
  //   else {
  //     return null;
  //   }
  // }, [definition, dimensions.width]);

  // if (definition.options && definition.options.display === "none") {
  //   return null;
  // }

  const baseStyles = classNames(
    "grid grid-cols-12 col-span-12",
    definition.width ? getResponsivePanelWidthClass(definition.width) : null,
    "overflow-auto"
  );

  const ErrorComponent = Error;
  const PlaceholderComponent = Placeholder.component;

  return (
    <PanelProvider
      definition={definition}
      dimensions={dimensions}
      showExpand={showExpand}
    >
      <div
        ref={panelRef}
        id={definition.name}
        className={baseStyles}
        // style={{ height: height ? `${height}px` : "auto" }}
      >
        <div className="col-span-12">
          <PlaceholderComponent
            animate={!!children}
            ready={ready || !!definition.error}
            // type={type}
          >
            <ErrorComponent error={definition.error} />
            <>{!definition.error ? children : null}</>
          </PlaceholderComponent>
        </div>
      </div>
    </PanelProvider>
  );
};

export default memo(Panel);
