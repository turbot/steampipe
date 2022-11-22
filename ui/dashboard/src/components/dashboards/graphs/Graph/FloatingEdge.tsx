import RowProperties, { RowPropertiesTitle } from "./RowProperties";
import Tooltip from "./Tooltip";
import { circleGetBezierPath, getEdgeParams } from "./utils";
import { classNames } from "../../../../utils/styles";
import { EdgeLabelRenderer, useStore } from "reactflow";
import { useCallback } from "react";
import { ThemeNames } from "../../../../hooks/useTheme";
import { useDashboard } from "../../../../hooks/useDashboard";

const FloatingEdge = ({
  id,
  source,
  target,
  markerEnd,
  style,
  data: {
    category,
    color,
    customColor,
    fields,
    labelOpacity,
    lineOpacity,
    row_data,
    label,
  },
}) => {
  const {
    themeContext: { theme },
  } = useDashboard();
  const sourceNode = useStore(
    useCallback((store) => store.nodeInternals.get(source), [source])
  );
  const targetNode = useStore(
    useCallback((store) => store.nodeInternals.get(target), [target])
  );

  if (!sourceNode || !targetNode) {
    return null;
  }

  const { sx, sy, mx, my, tx, ty, sourcePos, targetPos } = getEdgeParams(
    sourceNode,
    targetNode
  );

  const d = circleGetBezierPath({
    sourceX: sx,
    sourceY: sy,
    sourcePosition: sourcePos,
    targetPosition: targetPos,
    targetX: tx,
    targetY: ty,
  });

  const edgeLabel = (
    <span
      title={label}
      className={classNames(
        "block p-px italic max-w-[70px] text-sm text-center text-wrap leading-tight line-clamp-2",
        row_data && row_data.properties
          ? "-mt-1 underline decoration-dashed decoration-2 underline-offset-3 decoration-black-scale-3"
          : null,
        customColor && theme.name === ThemeNames.STEAMPIPE_DARK
          ? "brightness-[1.75]"
          : null
      )}
      style={{ color, opacity: labelOpacity }}
    >
      {label}
    </span>
  );

  const edgeLabelWrapper = (
    <div
      className={classNames(
        "bg-dashboard-panel",
        row_data && row_data.properties
          ? "flex space-x-0.5 items-center"
          : undefined
      )}
    >
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
    </div>
  );

  return (
    <>
      <path
        className={
          customColor && theme.name === ThemeNames.STEAMPIPE_DARK
            ? "brightness-[1.75]"
            : undefined
        }
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
