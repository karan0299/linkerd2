import React from 'react';
import _isFunction from 'lodash/isFunction';
import { usePageVisibility } from 'react-page-visibility';

export function handlePageVisibility(params) {
  const { prevVisibilityState, currentVisibilityState, onVisible, onHidden } = params;
  if (prevVisibilityState !== undefined && currentVisibilityState !== undefined) {
    if (prevVisibilityState && !currentVisibilityState && _isFunction(onHidden)) {
      onHidden();
    }

    if (!prevVisibilityState && currentVisibilityState && _isFunction(onVisible)) {
      onVisible();
    }
  }
}

export const withPageVisibility = WrappedComponent => {
  const Component = props => {
    const isVisible = usePageVisibility();
    return <WrappedComponent {...props} isPageVisible={isVisible} />;
  };

  return Component;
};
