// Author: Benjamin Stonestreet
// Created: 2024-02-02

(function attachMyPay(global) {
  const STATE = {
    config: null,
    modal: null,
    backdrop: null,
    statusText: null,
    spinnerWrap: null,
    spinnerText: null,
    badge: null,
    invoiceValue: null,
    amountValue: null,
    destinationValue: null,
    eventSource: null,
    isOpen: false,
    openedAtMs: 0,
    ignoreNextBackdropClick: false,
  };

  function ensureSpinnerAnimation() {
    const existing = document.getElementById('mypay-spinner-style');
    if (existing) {
      return;
    }

    const style = document.createElement('style');
    style.id = 'mypay-spinner-style';
    style.textContent = '@keyframes mypay-spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }';
    document.head.appendChild(style);
  }

  function normalizeBaseUrl(value) {
    if (!value || typeof value !== 'string') {
      return '';
    }

    return value.endsWith('/') ? value.slice(0, -1) : value;
  }

  function getApiUrl(path) {
    const base = normalizeBaseUrl(STATE.config?.apiBaseUrl || '');
    return `${base}${path}`;
  }

  function isDebugEnabled() {
    return !!STATE.config?.debug;
  }

  function isUuid(value) {
    return typeof value === 'string' && /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
  }

  function isHex64(value) {
    return typeof value === 'string' && /^[A-Fa-f0-9]{64}$/.test(value);
  }

  function sanitizeForDebug(value, depth, seen) {
    if (depth > 4) {
      return '[MaxDepth]';
    }

    if (value === null || value === undefined) {
      return value;
    }

    const valueType = typeof value;
    if (valueType === 'string') {
      if (value.length > 240) {
        return `${value.slice(0, 240)}...[truncated]`;
      }
      return value;
    }

    if (valueType === 'number' || valueType === 'boolean') {
      return value;
    }

    if (valueType !== 'object') {
      return `[${valueType}]`;
    }

    if (seen.has(value)) {
      return '[Circular]';
    }
    seen.add(value);

    if (Array.isArray(value)) {
      return value.slice(0, 30).map((item) => sanitizeForDebug(item, depth + 1, seen));
    }

    const sensitive = /secret|token|password|mnemonic|seed|private|signature|cookie|jwt|auth|key/i;
    const output = {};
    for (const [entryKey, entryValue] of Object.entries(value)) {
      if (sensitive.test(entryKey)) {
        output[entryKey] = '[Redacted]';
        continue;
      }

      output[entryKey] = sanitizeForDebug(entryValue, depth + 1, seen);
    }

    return output;
  }

  function debugLog(label, value) {
    if (!isDebugEnabled()) {
      return;
    }

    const sanitized = sanitizeForDebug(value, 0, new WeakSet());
    console.log(`[MyPay debug] ${label}`, sanitized);
  }

  function getCrossmarkClient() {
    const client = global.crossmark || global.Crossmark || null;
    if (!client || (typeof client !== 'object' && typeof client !== 'function')) {
      return null;
    }

    return client;
  }

  function resolvePathFunction(root, path) {
    let target = root;
    for (let index = 0; index < path.length - 1; index += 1) {
      const key = path[index];
      if (!target || (typeof target !== 'object' && typeof target !== 'function')) {
        return null;
      }
      target = target[key];
    }

    if (!target || (typeof target !== 'object' && typeof target !== 'function')) {
      return null;
    }

    const methodName = path[path.length - 1];
    const method = target[methodName];
    if (typeof method !== 'function') {
      return null;
    }

    return {
      target,
      method,
      label: path.join('.'),
    };
  }

  function detectCrossmarkInvoker(client) {
    const candidates = [
      ['signAndSubmit'],
      ['methods', 'signAndSubmit'],
      ['async', 'signAndSubmit'],
      ['sync', 'signAndSubmit'],
      ['methods', 'submitAndWait'],
      ['async', 'submitAndWait'],
      ['sync', 'submitAndWait'],
      ['sign'],
      ['methods', 'sign'],
      ['async', 'sign'],
      ['sync', 'sign'],
    ];

    for (const path of candidates) {
      const resolved = resolvePathFunction(client, path);
      if (resolved) {
        return resolved;
      }
    }

    return null;
  }

  function summarizeCrossmarkKeys(client) {
    if (!client || (typeof client !== 'object' && typeof client !== 'function')) {
      return 'none';
    }

    const topLevel = Object.keys(client).slice(0, 12).join(', ');
    const methodsLevel = client.methods && typeof client.methods === 'object'
      ? Object.keys(client.methods).slice(0, 12).join(', ')
      : '';
    const asyncLevel = client.async && typeof client.async === 'object'
      ? Object.keys(client.async).slice(0, 12).join(', ')
      : '';
    const syncLevel = client.sync && typeof client.sync === 'object'
      ? Object.keys(client.sync).slice(0, 12).join(', ')
      : '';

    const parts = [`top=[${topLevel || 'none'}]`];
    if (methodsLevel) {
      parts.push(`methods=[${methodsLevel}]`);
    }
    if (asyncLevel) {
      parts.push(`async=[${asyncLevel}]`);
    }
    if (syncLevel) {
      parts.push(`sync=[${syncLevel}]`);
    }

    return parts.join(' ');
  }

  async function ensureCrossmarkMounted(client) {
    if (!client || typeof client.mount !== 'function') {
      return;
    }

    try {
      const result = client.mount();
      if (result && typeof result.then === 'function') {
        await result;
      }
    } catch (_error) {
      // Ignore mount errors and continue with method discovery.
    }
  }

  async function getCrossmarkInvoker() {
    const client = getCrossmarkClient();
    if (!client) {
      return null;
    }

    let invoker = detectCrossmarkInvoker(client);
    if (invoker) {
      return { client, invoker };
    }

    await ensureCrossmarkMounted(client);
    invoker = detectCrossmarkInvoker(client);
    if (!invoker) {
      return null;
    }

    return { client, invoker };
  }

  async function invokeCrossmark(invoker, payload) {
    try {
      return await invoker.method.call(invoker.target, payload);
    } catch (firstError) {
      try {
        return await invoker.method.call(invoker.target, { txjson: payload });
      } catch (_secondError) {
        throw firstError;
      }
    }
  }

  function isCrossmarkAvailable() {
    const client = getCrossmarkClient();
    return !!(client && detectCrossmarkInvoker(client));
  }

  function getCrossmarkDiagnostic() {
    const client = getCrossmarkClient();
    if (!client) {
      return 'Crossmark object not found on window.';
    }

    const invoker = detectCrossmarkInvoker(client);
    if (!invoker) {
      return `Crossmark detected but no supported signing method found. Available keys: ${summarizeCrossmarkKeys(client)}`;
    }

    return `Crossmark method detected: ${invoker.label}`;
  }

  function setStatus(message, isError) {
    if (!STATE.statusText) {
      return;
    }

    STATE.statusText.textContent = message;
    STATE.statusText.style.color = isError ? '#b91c1c' : '#111827';
  }

  function setSpinnerVisible(visible) {
    if (!STATE.spinnerWrap) {
      return;
    }

    STATE.spinnerWrap.style.display = visible ? 'flex' : 'none';
  }

  function setSpinnerText(message) {
    if (!STATE.spinnerText) {
      return;
    }

    STATE.spinnerText.textContent = message;
  }

  function setBadge(label, tone) {
    if (!STATE.badge) {
      return;
    }

    const styles = {
      neutral: {
        background: '#f3f4f6',
        color: '#374151',
      },
      active: {
        background: '#dcfce7',
        color: '#166534',
      },
      pending: {
        background: '#fef3c7',
        color: '#92400e',
      },
      success: {
        background: '#bbf7d0',
        color: '#166534',
      },
      error: {
        background: '#fee2e2',
        color: '#991b1b',
      },
    };

    const selected = styles[tone] || styles.neutral;
    STATE.badge.textContent = label;
    STATE.badge.style.backgroundColor = selected.background;
    STATE.badge.style.color = selected.color;
  }

  function setPaymentDetails(invoice) {
    if (!invoice) {
      return;
    }

    if (STATE.invoiceValue) {
      STATE.invoiceValue.textContent = invoice.invoiceId || STATE.config?.invoiceId || '-';
    }

    if (STATE.amountValue) {
      STATE.amountValue.textContent = `${invoice.amountDrops || '-'} drops`;
    }

    if (STATE.destinationValue) {
      const destination = invoice.merchantAddress || '-';
      STATE.destinationValue.textContent = destination;
      STATE.destinationValue.title = destination;
    }
  }

  function applyPhase(phase, message) {
    switch (phase) {
      case 'loading':
        setBadge('Preparing', 'active');
        setSpinnerText('Loading payment details...');
        setSpinnerVisible(true);
        setStatus(message || 'Loading invoice details...', false);
        break;
      case 'signing':
        setBadge('Awaiting Signature', 'active');
        setSpinnerText('Waiting for wallet signature...');
        setSpinnerVisible(true);
        setStatus(message || 'Opening Crossmark for signature...', false);
        break;
      case 'verifying':
        setBadge('Verifying', 'active');
        setSpinnerText('Submitting for verification...');
        setSpinnerVisible(true);
        setStatus(message || 'Submitting transaction for verification...', false);
        break;
      case 'pending':
        setBadge('Pending Confirmation', 'pending');
        setSpinnerText('Waiting for on-ledger confirmation...');
        setSpinnerVisible(true);
        setStatus(message || 'Payment created. Waiting for confirmation...', false);
        break;
      case 'paid':
        setBadge('Payment Complete', 'success');
        setSpinnerVisible(false);
        setStatus(message || 'Payment complete.', false);
        break;
      case 'error':
        setBadge('Action Required', 'error');
        setSpinnerVisible(false);
        setStatus(message || 'Payment failed. Please try again.', true);
        break;
      default:
        setBadge('Ready', 'neutral');
        setSpinnerText('Processing payment...');
        setSpinnerVisible(false);
        setStatus(message || 'Ready to start payment flow.', false);
        break;
    }
  }

  function closeInvoiceEventStream() {
    if (!STATE.eventSource) {
      return;
    }

    try {
      STATE.eventSource.close();
    } catch (_error) {
      // no-op
    }

    STATE.eventSource = null;
  }

  function buildInvoiceEventsUrl(invoiceId) {
    const base = normalizeBaseUrl(STATE.config?.apiBaseUrl || '');
    return `${base}/api/invoices/${encodeURIComponent(invoiceId)}/events`;
  }

  function subscribeInvoiceEvents(invoiceId) {
    if (!invoiceId) {
      return;
    }

    if (typeof global.EventSource !== 'function') {
      debugLog('EventSource unavailable; cannot subscribe invoice updates', { invoiceId });
      return;
    }

    closeInvoiceEventStream();

    const eventsUrl = buildInvoiceEventsUrl(invoiceId);
    const source = new global.EventSource(eventsUrl);
    STATE.eventSource = source;
    debugLog('Subscribed to invoice events', { invoiceId, eventsUrl });

    source.onmessage = (event) => {
      if (!event?.data) {
        return;
      }

      try {
        const payload = JSON.parse(event.data);
        const status = (payload?.status || '').toLowerCase();
        if (!status) {
          return;
        }

        debugLog('Invoice status event received', payload);

        if (status === 'paid') {
          applyPhase('paid', 'Payment confirmed by backend.');
          closeInvoiceEventStream();

          if (STATE.config?.successUrl) {
            setTimeout(() => {
              global.location.assign(STATE.config.successUrl);
            }, 900);
          }
          return;
        }

        if (status === 'verification_failed') {
          applyPhase('error', 'Backend verification failed. Please retry payment.');
          closeInvoiceEventStream();
          return;
        }

        if (status === 'verification_pending' || status === 'created') {
          applyPhase('pending', 'Waiting for backend confirmation...');
        }
      } catch (_parseError) {
        debugLog('Failed to parse invoice status event', { raw: event.data });
      }
    };

    source.onerror = () => {
      debugLog('Invoice event stream disconnected', { invoiceId });
    };
  }

  function closeModal() {
    if (!STATE.modal || !STATE.backdrop) {
      return;
    }

    STATE.modal.style.display = 'none';
    STATE.backdrop.style.display = 'none';
    setSpinnerVisible(false);
    closeInvoiceEventStream();
    STATE.isOpen = false;
  }

  function openModal() {
    if (!STATE.modal || !STATE.backdrop) {
      throw new Error('MyPay modal is not initialized. Call MyPay.init first.');
    }

    STATE.openedAtMs = Date.now();
    STATE.ignoreNextBackdropClick = true;
    STATE.modal.style.display = 'block';
    STATE.backdrop.style.display = 'block';
    STATE.isOpen = true;
  }

  function buildModal() {
    if (STATE.modal && STATE.backdrop && STATE.statusText) {
      return;
    }

    ensureSpinnerAnimation();

    const backdrop = document.createElement('div');
    Object.assign(backdrop.style, {
      position: 'fixed',
      inset: '0',
      backgroundColor: 'rgba(15, 23, 42, 0.4)',
      zIndex: '10000',
      display: 'none',
    });

    const modal = document.createElement('div');
    Object.assign(modal.style, {
      position: 'fixed',
      top: '50%',
      left: '50%',
      transform: 'translate(-50%, -50%)',
      width: 'min(420px, calc(100vw - 32px))',
      borderRadius: '12px',
      border: '1px solid #e5e7eb',
      backgroundColor: '#ffffff',
      zIndex: '10001',
      display: 'none',
      boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
      fontFamily: 'system-ui, -apple-system, Segoe UI, Roboto, sans-serif',
    });

    const body = document.createElement('div');
    Object.assign(body.style, {
      padding: '20px',
      display: 'grid',
      gap: '12px',
    });

    const title = document.createElement('h2');
    title.textContent = 'Pay with XRPL';
    Object.assign(title.style, {
      margin: '0',
      fontSize: '20px',
      color: '#111827',
    });

    const status = document.createElement('p');
    status.textContent = 'Ready to start payment flow.';
    Object.assign(status.style, {
      margin: '0',
      fontSize: '14px',
      color: '#111827',
      lineHeight: '1.5',
    });

    const badge = document.createElement('span');
    badge.textContent = 'Ready';
    Object.assign(badge.style, {
      justifySelf: 'start',
      display: 'inline-flex',
      alignItems: 'center',
      padding: '4px 10px',
      borderRadius: '999px',
      fontSize: '12px',
      fontWeight: '600',
      backgroundColor: '#f3f4f6',
      color: '#374151',
    });

    const detailsWrap = document.createElement('div');
    Object.assign(detailsWrap.style, {
      border: '1px solid #e5e7eb',
      borderRadius: '10px',
      padding: '10px',
      display: 'grid',
      gap: '6px',
      backgroundColor: '#fafafa',
    });

    const createDetailRow = (labelText, valueText) => {
      const row = document.createElement('div');
      Object.assign(row.style, {
        display: 'grid',
        gap: '2px',
      });

      const label = document.createElement('span');
      label.textContent = labelText;
      Object.assign(label.style, {
        fontSize: '12px',
        color: '#6b7280',
      });

      const value = document.createElement('span');
      value.textContent = valueText;
      Object.assign(value.style, {
        fontSize: '13px',
        color: '#111827',
        fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
        wordBreak: 'break-all',
      });

      row.append(label, value);
      return { row, value };
    };

    const invoiceRow = createDetailRow('Invoice', STATE.config?.invoiceId || '-');
    const amountRow = createDetailRow('Amount', '-');
    const destinationRow = createDetailRow('Destination', '-');
    detailsWrap.append(invoiceRow.row, amountRow.row, destinationRow.row);

    const spinnerWrap = document.createElement('div');
    Object.assign(spinnerWrap.style, {
      display: 'none',
      alignItems: 'center',
      gap: '8px',
    });

    const spinner = document.createElement('div');
    Object.assign(spinner.style, {
      width: '16px',
      height: '16px',
      borderRadius: '9999px',
      border: '2px solid #86efac',
      borderTopColor: '#16a34a',
      animation: 'mypay-spin 0.8s linear infinite',
      flexShrink: '0',
    });

    const spinnerText = document.createElement('span');
    spinnerText.textContent = 'Processing payment...';
    Object.assign(spinnerText.style, {
      fontSize: '13px',
      color: '#15803d',
    });

    spinnerWrap.append(spinner, spinnerText);

    const closeButton = document.createElement('button');
    closeButton.type = 'button';
    closeButton.textContent = 'Close';
    Object.assign(closeButton.style, {
      justifySelf: 'end',
      padding: '8px 12px',
      borderRadius: '8px',
      border: '1px solid #d1d5db',
      backgroundColor: '#ffffff',
      cursor: 'pointer',
      color: '#111827',
    });

    closeButton.addEventListener('click', closeModal);
    backdrop.addEventListener('click', (event) => {
      if (event.target !== backdrop) {
        return;
      }

      if (STATE.ignoreNextBackdropClick) {
        STATE.ignoreNextBackdropClick = false;
        return;
      }

      if (Date.now() - STATE.openedAtMs < 400) {
        return;
      }

      closeModal();
    });

    body.append(title, badge, detailsWrap, status, spinnerWrap, closeButton);
    modal.appendChild(body);
    document.body.append(backdrop, modal);

    STATE.modal = modal;
    STATE.backdrop = backdrop;
    STATE.statusText = status;
    STATE.spinnerWrap = spinnerWrap;
    STATE.spinnerText = spinnerText;
    STATE.badge = badge;
    STATE.invoiceValue = invoiceRow.value;
    STATE.amountValue = amountRow.value;
    STATE.destinationValue = destinationRow.value;
  }

  async function fetchInvoiceDetails(invoiceId) {
    const response = await fetch(getApiUrl(`/api/invoices/${encodeURIComponent(invoiceId)}`), {
      method: 'GET',
      headers: { Accept: 'application/json' },
    });

    if (!response.ok) {
      throw new Error(`Invoice lookup failed (${response.status})`);
    }

    return response.json();
  }

  function resolveTxHash(result) {
    if (isHex64(result) || isUuid(result)) {
      debugLog('Resolved tx reference from direct string response', { txReference: result });
      return result;
    }

    if (!result || typeof result !== 'object') {
      return '';
    }

    const knownPathCandidates = [
      result.tx_hash,
      result.txHash,
      result.transactionHash,
      result.hash,
      result.txid,
      result.id,
      result?.response?.tx_hash,
      result?.response?.txHash,
      result?.response?.hash,
      result?.response?.txid,
      result?.response?.transactionHash,
      result?.response?.result?.hash,
      result?.response?.result?.tx_hash,
      result?.response?.result?.txHash,
      result?.response?.result?.txid,
      result?.data?.tx_hash,
      result?.data?.txHash,
      result?.data?.hash,
      result?.data?.txid,
      result?.data?.transactionHash,
      result?.data?.result?.hash,
      result?.result?.hash,
      result?.result?.tx_hash,
      result?.result?.txHash,
      result?.result?.txid,
      result?.meta?.tx_hash,
      result?.meta?.hash,
      result?.payload?.tx_hash,
      result?.payload?.hash,
    ];

    for (const candidate of knownPathCandidates) {
      if (isHex64(candidate) || isUuid(candidate)) {
        debugLog('Resolved tx reference from known path', { txReference: candidate });
        return candidate;
      }
    }

    const visited = new WeakSet();
    const stack = [result];
    let traversed = 0;

    while (stack.length > 0 && traversed < 2500) {
      const current = stack.pop();
      traversed += 1;

      if (!current || typeof current !== 'object') {
        continue;
      }

      if (visited.has(current)) {
        continue;
      }
      visited.add(current);

      for (const value of Object.values(current)) {
        if (isHex64(value) || isUuid(value)) {
          debugLog('Resolved tx reference from recursive scan', { txReference: value });
          return value;
        }

        if (value && typeof value === 'object') {
          stack.push(value);
        }
      }
    }

    debugLog('Failed to resolve tx hash from response', result);
    return '';
  }

  async function submitVerify(invoiceId, txHash) {
    const response = await fetch(getApiUrl('/api/verify'), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({
        invoice_id: invoiceId,
        tx_hash: txHash,
      }),
    });

    if (response.status !== 202 && response.status !== 200) {
      throw new Error(`Verify request failed (${response.status})`);
    }
  }

  function buildCrossmarkPayload(invoice) {
    return {
      TransactionType: 'Payment',
      Destination: invoice.merchantAddress,
      DestinationTag: invoice.destinationTag,
      Amount: String(invoice.amountDrops),
    };
  }

  async function runCheckoutFlow() {
    const invoiceId = STATE.config?.invoiceId;
    if (!invoiceId) {
      throw new Error('Missing invoiceId in MyPay.init config');
    }

    const crossmark = await getCrossmarkInvoker();
    if (!crossmark) {
      throw new Error(`Crossmark extension is required. ${getCrossmarkDiagnostic()}`);
    }

    applyPhase('loading');
    const invoice = await fetchInvoiceDetails(invoiceId);
    setPaymentDetails(invoice);

    applyPhase('signing');
    const payload = buildCrossmarkPayload(invoice);
    debugLog('Crossmark payload', payload);
    const signResult = await invokeCrossmark(crossmark.invoker, payload);
    debugLog('Crossmark response', signResult);

    const txHash = resolveTxHash(signResult);
    if (!txHash) {
      throw new Error('Unable to read transaction hash from Crossmark response.');
    }

    applyPhase('verifying');
    await submitVerify(invoiceId, txHash);

    if (isUuid(txHash)) {
      subscribeInvoiceEvents(invoiceId);
      applyPhase('pending', 'Crossmark request created. Waiting for backend confirmation.');
      return;
    }

    applyPhase('paid', 'Payment complete. Finalizing...');
    if (STATE.config?.successUrl) {
      setTimeout(() => {
        global.location.assign(STATE.config.successUrl);
      }, 900);
    }
  }

  async function startCheckoutFlow() {
    openModal();

    try {
      await runCheckoutFlow();
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unexpected error during checkout.';
      applyPhase('error', message);
      throw error;
    }
  }

  function bindTrigger() {
    const triggerElement = document.querySelector(STATE.config.triggerSelector);
    if (!triggerElement) {
      throw new Error(`No element found for triggerSelector: ${STATE.config.triggerSelector}`);
    }

    triggerElement.addEventListener('click', async (event) => {
      event.preventDefault();
      event.stopPropagation();
      try {
        await startCheckoutFlow();
      } catch (_error) {
        // Handled by startCheckoutFlow status updates.
      }
    });
  }

  function validateConfig(config) {
    if (!config || typeof config !== 'object') {
      throw new Error('MyPay.init requires a configuration object.');
    }

    if (!config.invoiceId || typeof config.invoiceId !== 'string') {
      throw new Error('MyPay.init requires invoiceId (string).');
    }

    if (!config.triggerSelector || typeof config.triggerSelector !== 'string') {
      throw new Error('MyPay.init requires triggerSelector (string).');
    }
  }

  const MyPay = {
    init(config) {
      validateConfig(config);

      STATE.config = {
        invoiceId: config.invoiceId,
        triggerSelector: config.triggerSelector,
        successUrl: config.successUrl,
        apiBaseUrl: config.apiBaseUrl || '',
        debug: !!config.debug,
      };

      buildModal();
      bindTrigger();
      applyPhase('idle', 'Ready to start payment flow.');
      if (STATE.invoiceValue) {
        STATE.invoiceValue.textContent = STATE.config.invoiceId;
      }
      if (isCrossmarkAvailable()) {
        setStatus('Crossmark detected. Ready to start payment flow.', false);
      } else {
        setStatus('Crossmark will be checked when payment starts. If unavailable, enable extension and refresh.', false);
      }
      setSpinnerVisible(false);
      return MyPay;
    },
    open() {
      openModal();
    },
    async start() {
      await startCheckoutFlow();
    },
    setInvoiceId(nextInvoiceId) {
      if (!nextInvoiceId || typeof nextInvoiceId !== 'string') {
        throw new Error('setInvoiceId requires a non-empty string invoice id.');
      }

      if (!STATE.config) {
        throw new Error('Call MyPay.init before setInvoiceId.');
      }

      STATE.config.invoiceId = nextInvoiceId;
      if (STATE.invoiceValue) {
        STATE.invoiceValue.textContent = nextInvoiceId;
      }
      closeInvoiceEventStream();
      return MyPay;
    },
    updateStatus(status, message) {
      if (typeof status !== 'string') {
        throw new Error('updateStatus requires a status string.');
      }

      const normalized = status.trim().toLowerCase();
      if (!normalized) {
        throw new Error('updateStatus requires a non-empty status string.');
      }

      const supported = {
        idle: 'idle',
        loading: 'loading',
        signing: 'signing',
        verifying: 'verifying',
        pending: 'pending',
        paid: 'paid',
        error: 'error',
      };

      const phase = supported[normalized];
      if (!phase) {
        throw new Error(`Unsupported status: ${status}`);
      }

      applyPhase(phase, message);
      return MyPay;
    },
    markPaid(message) {
      applyPhase('paid', message || 'Payment confirmed by backend.');
      return MyPay;
    },
    close() {
      closeModal();
    },
  };

  global.MyPay = MyPay;
})(window);