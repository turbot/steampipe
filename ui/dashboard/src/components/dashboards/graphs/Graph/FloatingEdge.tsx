import RowProperties, { RowPropertiesTitle } from "./RowProperties";
import Tooltip from "./Tooltip";
import {
  buildLabelTextShadow,
  circleGetBezierPath,
  getEdgeParams,
} from "./utils";
import { classNames } from "../../../../utils/styles";
import { colorToRgb } from "../../../../utils/color";
import { EdgeLabelRenderer, useStore } from "reactflow";
import { useCallback } from "react";

const FloatingEdge = ({
  id,
  source,
  target,
  markerEnd,
  style,
  data: {
    category,
    color,
    fields,
    labelOpacity,
    lineOpacity,
    row_data,
    label,
    themeColors,
  },
}) => {
  const sourceNode = useStore(
    useCallback((store) => store.nodeInternals.get(source), [source])
  );
  const targetNode = useStore(
    useCallback((store) => store.nodeInternals.get(target), [target])
  );

  if (!sourceNode || !targetNode) {
    return null;
  }

  const { sx, sy, mx, my, tx, ty } = getEdgeParams(sourceNode, targetNode);

  const d = circleGetBezierPath({
    sourceX: sx,
    sourceY: sy,
    targetX: tx,
    targetY: ty,
  });

  const colorRgb = colorToRgb(color, themeColors);

  const edgeLabel = (
    <span
      title={label}
      className={classNames(
        "block italic max-w-[70px] text-sm text-center text-wrap leading-tight line-clamp-2",
        row_data && row_data.properties ? "border-b border-dashed" : null
      )}
      style={{
        borderColor: `rgba(${colorRgb[0]},${colorRgb[1]},${colorRgb[2]},${labelOpacity})`,
        color: `rgba(${colorRgb[0]},${colorRgb[1]},${colorRgb[2]},${labelOpacity})`,
        textDecorationColor: `rgba(${colorRgb[0]},${colorRgb[1]},${colorRgb[2]},${labelOpacity})`,
        textShadow: buildLabelTextShadow(themeColors.dashboardPanel),
      }}
    >
      {label}
    </span>
  );

  const edgeLabelWrapper = (
    <>
      {row_data && row_data.properties && (
        <Tooltip
          overlay={
            <RowProperties
              fields={fields || null}
              properties={row_data.properties}
            />
          }
          title={<RowPropertiesTitle category={category} title={label} />}
        >
          {edgeLabel}
        </Tooltip>
      )}
      {(!row_data || !row_data.properties) && edgeLabel}
    </>
  );

  return (
    <>
      <path
        id={id}
        d={d}
        markerEnd={markerEnd}
        style={{
          ...(style || {}),
          opacity: lineOpacity,
          stroke: color,
          strokeWidth: 1,
        }}
      />
      <EdgeLabelRenderer>
        <div
          className="absolute pointer-events-auto cursor-grab"
          style={{
            transform: `translate(-50%, -50%) translate(${mx}px,${my}px)`,
          }}
        >
          {edgeLabelWrapper}
        </div>
      </EdgeLabelRenderer>
    </>
  );
};

export default FloatingEdge;
