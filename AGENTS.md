# AIエージェントへの指示

## 開始時

- [`dev-platform`](https://github.com/kikudesuyo/dev-platform) の最新 `AGENTS.md`、`README.md`、`dev-guideline/README.md`、`dev-guideline/AI_AGENT_COMPLETION.md` を確認する。
- Goコードを変更する場合は [`go-api/BLUEPRINT.md`](https://github.com/kikudesuyo/dev-platform/blob/main/go-api/BLUEPRINT.md) を確認する。
- IssueとPull Requestは1対1にし、Issue用branch `issue/<number>` とworktreeで作業する。
- 既存の差分を破棄・上書きしない。

## Go規約

- snake_caseのファイル名、CamelCaseの型・関数名を使用する。
- handler/entityは機能単位、service/repositoryは責務単位でファイルを分ける。
- handler、service、repositoryの責務を混ぜない。
- context.Contextをリクエスト境界から下位層へ伝播する。
- スライス変数は複数形でなく `List` 接尾辞を使用する。
- エラーを無視せず、機密値・イベント本文をログに出さない。
- API JSONキーはBuildlogとの契約に定義されたlower camel caseを使用する。

## 機密情報

- 実資格情報をソース、テストfixture、ログへ書かない。
- Cloud RunではSecret Manager連携を使用する。
- GitHub Actionsで機密値が必要になった場合はWIFとSecret Managerを使用し、GitHub Secretsを参照しない。
