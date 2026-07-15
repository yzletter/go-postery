# gRPC Test Commands

## User Avatar Callback

直传 OSS 成功后，可以直接调用 User 微服务的 `UploadAvatarCallback` 模拟 OSS 回调落库。

前提：

- User gRPC 服务已启动，默认测试地址是 `localhost:9007`。
- `ObjectName` 必须是 `users/avatar/{user_id}/{filename}` 格式。
- `UserID` 必须和 `ObjectName` 中的 `{user_id}` 一致。

```bash
USER_GRPC_ADDR=localhost:9007
AVATAR_OBJECT=users/avatar/1/avatar-test.png

grpcurl -plaintext \
  -import-path . \
  -proto api/proto/user/v1/user.proto \
  -d "{\"UserID\":1,\"ObjectName\":\"${AVATAR_OBJECT}\"}" \
  "${USER_GRPC_ADDR}" \
  user.v1.UserService/UploadAvatarCallback
```

成功时返回空对象：

```json
{}
```

再查用户资料确认头像对象名已经落库：

```bash
grpcurl -plaintext \
  -import-path . \
  -proto api/proto/user/v1/user.proto \
  -d '{"ID":1}' \
  "${USER_GRPC_ADDR}" \
  user.v1.UserService/GetProfileById
```

返回结果里的 `Avatar` 应该是刚才传入的对象名，例如：

```json
{
  "UserID": "1",
  "Avatar": "users/avatar/1/avatar-test.png"
}
```
