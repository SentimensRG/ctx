# mergectx

Package `mergectx` provides utilities for merging `context.Context` objects such
that the resulting child context contains the union of the parent context's
values.

Functions `Join` and `Link` behave analogously to their counterparts in the `ctx`
package.
