import { useEffect, useState } from "react";

function usePaginatedList<T>(
  items: T[] = [],
  pageSize = 10,
  doubleOnExpand = true
) {
  const [allItems, setAllItems] = useState(items);
  const [visibleItems, setVisibleItems] = useState(items.slice(0, pageSize));
  const [nextPageSize, setNextPageSize] = useState(
    doubleOnExpand ? pageSize * 2 : pageSize
  );

  useEffect(() => {
    setAllItems(items);
    setVisibleItems(items.slice(0, pageSize));
    setNextPageSize(doubleOnExpand ? pageSize * 2 : pageSize);
  }, [doubleOnExpand, items, pageSize]);

  const hasMore = visibleItems.length < allItems.length;

  const loadMore = () => {
    if (!hasMore) {
      return;
    }
    const nextItems = allItems.slice(
      visibleItems.length,
      nextPageSize + visibleItems.length - 1
    );
    setVisibleItems([...visibleItems, ...nextItems]);
    setNextPageSize(doubleOnExpand ? nextPageSize * 2 : pageSize);
  };

  return { visibleItems, hasMore, loadMore };
}

export default usePaginatedList;
