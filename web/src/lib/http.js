// Extract a human-readable message from an axios error, falling back to a
// caller-supplied default when the server did not send one.
export function errorMessage(err, fallback) {
	return err?.response?.data?.error || fallback;
}
