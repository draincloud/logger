# Logger
## Usage 

```go
import (
    "log/slog"
    "github.com/draincloud/logger"
)

func someFunc(ctx context.Context, userID int64) {
    logger.Info(ctx, "someFunc call", slog.Int64("userID", userID))
    // ...
}

func someErrFunc(ctx context.Context) error {
    err := failingFunc(ctx)
    if err != nil {
        logger.Error(ctx, "someErrFunc error", logger.Err(err))
        return err
    }
    // ...
}
```