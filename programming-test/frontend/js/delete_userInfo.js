// テーブルの各行の「削除ボタン」が押された時の処理
// 指定の行のユーザー情報を削除
document.getElementById('userTable_body').addEventListener("click", function(event) {

    // フォームのデフォルトの送信を防止(js側でPOST送信するので)
    event.preventDefault();

    // テーブルに隠しパラメータとして置いているDB側の一意のIDを取得
    // ボタンの右隣を一意のidセルとしているので、その値を取得
    // 今回はテストなので決め打ちで取得するが、本来は定数か列挙型でどこかで順番を定義した方がよい
    var deleteCell = event.target.closest("td");
    var unique_id = deleteCell.nextElementSibling.textContent;

    // バックエンド側に一意のIDを渡すデータ準備
    // 渡すデータは1つなのでFormDataで送ってもいいし、JSONで送ってもいい
    var formData = new FormData();
    formData.append("ID", unique_id);

     // バックエンド側へPOSTリクエストを送信
     // ID情報を渡して、そのIDを持つレコードをバックエンド側に削除してもらう
     fetch("http://localhost:8081/deleteInfo",{
        method: "POST",
        body: formData,
    })
    .then(response => {
        // ユーザー情報が1件も登録されていない状態なら、デフォで1件追加する処理を組んである
        // その通知をアラートで行う
        if(!response.ok){
            alert("削除処理のレスポンスでエラー発生");
        } else if(response.status == 201){
            alert("テーブルに1件もユーザー情報が登録されていないため、\nデフォルトのユーザー情報を登録しました。")
            window.location.href = '/index.html';
        } else{
            // 削除処理が完了したので、Indexページへリダイレクト
            window.location.href = '/index.html';
        }
    });
});