import Icon from "../../../Icon";
import LoadingIndicator from "../../LoadingIndicator";
import Primitive from "../../Primitive";
import { tagIcon, tagsIcon } from "../../../../constants/icons";
import { useMemo } from "react";

const ResourceTags = ({ data, error }) => {
  const tags = useMemo(() => {
    if (error || !data) {
      return null;
    }

    if (data.length < 2) {
      return null;
    }

    const tagRow = data[1][0];

    if (!tagRow) {
      return null;
    }

    const tags = [];
    for (const [key, value] of Object.entries(tagRow)) {
      tags.push({ key, value });
    }
    return tags;
  }, [data, error]);

  return (
    <Primitive error={error} ready={true}>
      <table className="table-auto divide-y divide-table-divide border border-table-border overflow-hidden">
        <thead className="bg-table-head text-table-head">
          <tr>
            <th className="pl-4 pr-0 py-3 text-left text-xs font-normal tracking-wider whitespace-nowrap">
              <Icon icon={tagsIcon} />
            </th>
            <th className="px-4 sm:px-6 py-3 text-left text-sm font-normal tracking-wider whitespace-nowrap">
              Tags
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-table-divide">
          {!data && (
            <tr>
              <td className="pl-4 pr-0 py-3 align-top content-center text-sm text-gray-500 whitespace-nowrap">
                <LoadingIndicator />
              </td>
              <td className="px-4 sm:px-6 py-3 align-top content-center text-sm italic text-gray-500 whitespace-nowrap">
                Loading...
              </td>
            </tr>
          )}
          {data && !tags && (
            <tr>
              <td className="pl-4 pr-0 py-3 align-top content-center text-sm text-gray-500 whitespace-nowrap">
                <Icon icon={tagIcon} />
              </td>
              <td className="px-4 sm:px-6 py-3 align-top content-center text-sm italic text-gray-500 whitespace-nowrap">
                No tags
              </td>
            </tr>
          )}
          {tags &&
            tags.map((tag) => (
              <tr key={tag.key}>
                <td className="pl-4 pr-0 py-3 align-top content-center text-sm whitespace-nowrap">
                  <Icon icon={tagIcon} />
                </td>
                <td className="px-4 sm:px-6 py-3 align-top content-center text-sm whitespace-nowrap">
                  <span className="font-light text-table-header mr-2">
                    {tag.key}
                  </span>
                  <span className="font-light text-table-header mr-2">=</span>
                  <span className="font-mono">{tag.value}</span>
                </td>
              </tr>
            ))}
        </tbody>
      </table>
    </Primitive>
  );
};

export default {
  type: "resource_tags",
  component: ResourceTags,
};
