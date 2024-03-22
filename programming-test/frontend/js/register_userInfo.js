// ユーザー情報の登録処理
// addページの登録ボタンを押したときのイベントで起動
// バックエンドにフォームデータを渡して、データベースに情報を登録
document.getElementById("registrationForm").addEventListener("submit", function(event) {

    // フォームのデフォルトの送信を防止(js側でPOST送信するので)
    event.preventDefault();

    // html側からフォーム情報を取得
    var form = document.getElementById("registrationForm");
    if(!form){
        alert("フォーム情報の取得に失敗");
    }
    else{

        // 本来ならここで変な値が入力されていないか等の要素に沿った内容のチェックを行う
        // SQLインジェクションとかもここでやるかな
        // 今回はHTML側で文字数上限やtypeで制限しているので、ここで厳密なチェックは行わない

        // FormDataオブジェクトを作成
        var formData = new FormData(form);
        
        // バックエンド側へPOSTリクエストを送信
        fetch("http://localhost:8081/registerInfo",{
            method: "POST",
            body:formData
        })
        .then(response => {
            if(!response.ok){
                alert("登録処理のレスポンスでエラー発生");
            } else {
                // 登録処理が完了したので、Indexページへリダイレクト
                window.location.href = '/index.html';
            }
        });
    }
});