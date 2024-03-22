// データベースに登録された情報のテーブルを作成
// ページ読み込み準備完了のDOMContentLoadedタイミングで動作
document.addEventListener('DOMContentLoaded', function() {
    
    // ユーザー情報を表示する処理
    var tableBody = document.getElementById('userTable_body');
    if (tableBody) {
        fetch('http://localhost:8081/readInfo') // バックエンドのAPIエンドポイント
            .then(response => response.json())
            .then(users => {
                var userNo = 1; // 表に数字を振るためのカウント

                // DBに登録されている全データを読み出し
                users.forEach(user => {
                    const row = tableBody.insertRow();

                    // 各種セルに情報を設定
                    SetCellInfo(row.insertCell(0),userNo,"cell_no");            // No.(削除があった場合は詰められる)
                    SetCellInfo(row.insertCell(1),user.name,"cell_name");       // 名前
                    SetCellInfo(row.insertCell(2),user.age + "歳","cell_age");  // 年齢

                    // 作成日時
                    // 1行で表示すると長いので、日付と時間で2行表示
                    var fhalf = user.date.substr(0,10);     // 2024-01-01
                    var shalf = user.date.substr(11,18);    // 17:40:55
                    var dateCell = row.insertCell(3);
                    dateCell.innerHTML = fhalf + "<br>" + shalf;    // 日付と時間の間に改行を挟む
                    dateCell.id = "cell_date";
                    
                    // 削除ボタン
                    // このタイミングで表に一緒に付けておく
                    var buttonCell = row.insertCell(4);
                    buttonCell.innerHTML = '<button type="submit" class="deleteButton">削除</button>';

                    // データベースに登録された一意のIDも覚えておきたいので隠し属性で書いておく
                    var idCell = row.insertCell(5);
                    idCell.className = "hiddenParam";               // class名設定して、cssで隠す
                    SetCellInfo(idCell,user.id, "cell_id");

                    // 次の行へ
                    userNo++;
                });
            })
            .catch(error => console.error('Error:', error));
    }
});



// 指定のセルにセル値とid(htmlの)を設定
function SetCellInfo(cell, value, id){
    cell.textContent = value
    cell.id = id
}