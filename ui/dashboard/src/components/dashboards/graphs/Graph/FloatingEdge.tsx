import Icon from "../../../Icon";
import RowProperties from "./RowProperties";
import Tooltip from "./Tooltip";
import { circleGetBezierPath, getEdgeParams } from "./utils";
import { EdgeLabelRenderer, useStore } from "reactflow";
import { useCallback } from "react";
import { classNames } from "../../../../utils/styles";

const FloatingEdge = ({
  id,
  source,
  target,
  markerEnd,
  style,
  data: { color, fields, row_data, label, themeColors },
}) => {
  // const edgeLabelRef = useRef(null);
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

  //const d = getSimpleBezierPath({
  const d = circleGetBezierPath({
    sourceX: sx,
    sourceY: sy,
    sourcePosition: sourcePos,
    targetPosition: targetPos,
    targetX: tx,
    targetY: ty,
  });

  const edgePropertiesIcon =
    row_data && row_data.properties ? (
      <Tooltip
        overlay={
          <RowProperties
            fields={fields || null}
            properties={row_data.properties}
          />
        }
        title={label}
      >
        <div className="cursor-pointer text-black-scale-3">
          <Icon className="w-3 h-3 rotate-90" icon="view-columns" />
        </div>
      </Tooltip>
    ) : null;

  const edgeLabel = (
    <div
      className={
        row_data && row_data.properties
          ? "flex space-x-0.5 items-center"
          : undefined
      }
    >
      <span
        title={label}
        className={classNames(
          "block p-px bg-dashboard-panel text-black-scale-4 italic max-w-[70px] text-sm text-center text-wrap line-clamp-2",
          row_data && row_data.properties
            ? "-mt-1 underline decoration-dashed decoration-2 underline-offset-2 decoration-black-scale-3 cursor-pointer"
            : null
        )}
      >
        {label}
      </span>
      {/*{edgePropertiesIcon}*/}
    </div>
  );

  return (
    <>
      <path
        id={id}
        d={d}
        markerEnd={markerEnd}
        style={{
          ...(style || {}),
          stroke: color || themeColors.blackScale4,
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
          {edgeLabel}
        </div>
      </EdgeLabelRenderer>
    </>
  );
};

export default FloatingEdge;
