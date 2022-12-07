import isEmpty from "lodash/isEmpty";
import useChartThemeColors from "../../../../hooks/useChartThemeColors";
import useDeepCompareEffect from "use-deep-compare-effect";
import useTemplateRender from "../../../../hooks/useTemplateRender";
import {
  Category,
  CategoryFields,
  KeyValuePairs,
  RowRenderResult,
} from "../../common/types";
import { classNames } from "../../../../utils/styles";
import { DashboardDataModeLive } from "../../../../types";
import { ErrorIcon } from "../../../../constants/icons";
import { getColorOverride } from "../../common";
import { isRelativeUrl } from "../../../../utils/url";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useEffect, useState } from "react";

interface RowPropertiesTitleProps {
  category: Category | undefined;
  title: string;
}

interface RowPropertiesProps {
  fields: CategoryFields | null;
  properties: KeyValuePairs | null;
}

interface RowPropertyItemProps {
  name: string;
  rowTemplateData?: RowRenderResult | null;
  value: any;
  wrap: boolean;
}

const RowPropertiesTitle = ({ category, title }: RowPropertiesTitleProps) => {
  const themeColors = useChartThemeColors();
  return (
    <div className="flex flex-col space-y-1">
      {category && (
        <span
          className="block text-foreground-lighter text-xs"
          style={{ color: getColorOverride(category.color, themeColors) }}
        >
          {category.title || category.name}
        </span>
      )}
      <strong className="block">{title}</strong>
    </div>
  );
};

const RowPropertyItemValue = ({
  name,
  rowTemplateData,
  value,
  wrap,
}: RowPropertyItemProps) => {
  const {
    components: { ExternalLink },
    dataMode,
  } = useDashboard();
  const [href, setHref] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  useEffect(() => {
    if (!rowTemplateData) {
      setHref(null);
      setError(null);
      return;
    }
    const renderedTemplateForField = rowTemplateData[name];
    if (!renderedTemplateForField) {
      setHref(null);
      setError(null);
      return;
    }
    if (renderedTemplateForField.result) {
      // We only want to render the HREF if it's live, or it's snapshot and absolute
      const isRelative = isRelativeUrl(renderedTemplateForField.result);
      setHref(
        dataMode !== DashboardDataModeLive && isRelative
          ? null
          : renderedTemplateForField.result
      );
      setError(null);
    } else if (renderedTemplateForField.error) {
      setHref(null);
      setError(renderedTemplateForField.error);
    }
  }, [dataMode, name, rowTemplateData]);

  const wrapClassName = wrap ? "whitespace-normal" : "truncate";
  const linkClassName = classNames("link-highlight", wrapClassName);

  let cellContent;
  if (value === null || value === undefined) {
    cellContent = href ? (
      <ExternalLink to={href} className={linkClassName} title={`${name}=null`}>
        null
      </ExternalLink>
    ) : (
      <span
        className={classNames("text-foreground-lightest", wrapClassName)}
        title={`${name}=null`}
      >
        <>null</>
      </span>
    );
  } else {
    let renderValue: string;
    switch (typeof value) {
      case "object":
        renderValue = JSON.stringify(value, null, 2);
        break;
      default:
        renderValue = value.toString();
        break;
    }
    cellContent = href ? (
      <ExternalLink
        to={href}
        className={linkClassName}
        title={`${name}=${renderValue ? renderValue : "Empty"}`}
      >
        {renderValue ? renderValue : "Empty"}
      </ExternalLink>
    ) : (
      <span
        className={classNames(
          "block break-words",
          renderValue ? "" : "text-foreground-lightest",
          wrapClassName
        )}
        title={`${name}=${renderValue ? renderValue : "Empty"}`}
      >
        {renderValue ? renderValue : "Empty"}
      </span>
    );
  }

  return error ? (
    <span className="flex items-center space-x-2" title={error}>
      {cellContent} <ErrorIcon className="inline h-4 w-4 text-alert" />
    </span>
  ) : (
    cellContent
  );
};

const RowPropertyItem = ({
  name,
  value,
  rowTemplateData,
  wrap,
}: RowPropertyItemProps) => {
  return (
    <div className="w-full">
      <span
        className={classNames(
          "block text-sm text-foreground-lighter truncate",
          wrap ? "whitespace-normal" : "truncate"
        )}
        title={name}
      >
        {name}
      </span>
      <RowPropertyItemValue
        name={name}
        rowTemplateData={rowTemplateData}
        wrap={wrap}
        value={value}
      />
    </div>
    // <div className="space-x-2">
    //   <span>{name}</span>
    //   <span>=</span>
    //   <span>{value}</span>
    // </div>
  );
};

const RowProperties = ({ properties = {}, fields }: RowPropertiesProps) => {
  const [rowTemplateData, setRowTemplateData] =
    useState<RowRenderResult | null>(null);
  const { ready: templateRenderReady, renderTemplates } = useTemplateRender();

  useDeepCompareEffect(() => {
    if (!templateRenderReady || isEmpty(fields) || !properties) {
      setRowTemplateData(null);
      return;
    }

    const doRender = async () => {
      const templates = {};
      for (const [name, field] of Object.entries(fields || {})) {
        if (field.display !== "none" && field.href) {
          templates[name] = field.href;
        }
      }
      // const templates = Object.fromEntries(
      //   fields
      //     .filter((col) => col.display !== "none" && !!col.href_template)
      //     .map((col) => [col.name, col.href_template as string])
      // );
      if (isEmpty(templates) || !properties) {
        setRowTemplateData(null);
        return;
      }
      const renderedResults = await renderTemplates(templates, [properties]);
      setRowTemplateData(renderedResults[0]);
    };

    doRender();
  }, [fields, properties, renderTemplates, templateRenderReady]);

  return (
    <div className="space-y-2">
      {Object.entries(properties || {}).map(([key, value]) => {
        const fieldDefinition = fields?.[key];
        if (fieldDefinition && fieldDefinition.display === "none") {
          return null;
        }
        return (
          <RowPropertyItem
            key={key}
            name={key}
            value={value}
            rowTemplateData={rowTemplateData}
            wrap={fieldDefinition ? fieldDefinition.wrap === "all" : false}
          />
        );
      })}
    </div>
  );
};

export default RowProperties;

export { RowPropertiesTitle };
