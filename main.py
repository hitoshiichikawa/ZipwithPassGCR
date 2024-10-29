import os
import tempfile
import pyminizip
from flask import Flask, request, jsonify

def create_zip(request):
    # クエリパラメータからパスワードとzipファイル名を取得
    password = request.args.get('password')
    zip_filename = request.args.get('zip_filename', 'protected.zip')
    
    if not password:
        return jsonify({'error': 'Password is required'}), 400

    # 一時ディレクトリを作成
    with tempfile.TemporaryDirectory() as temp_dir:
        # POSTデータからファイルを取得
        files = request.files
        if not files:
            return jsonify({'error': 'No files uploaded'}), 400

        # 各ファイルを一時ディレクトリに保存
        for filename, file in files.items():
            file_path = os.path.join(temp_dir, filename)
            file.save(file_path)

        # パスワード付きzipファイルを作成
        zip_path = os.path.join(temp_dir, zip_filename)
        pyminizip.compress_multiple(
            [os.path.join(temp_dir, f) for f in files.keys()],
            [],
            zip_path,
            password,
            0
        )

        # zipファイルをメモリに読み込む
        with open(zip_path, 'rb') as f:
            zip_data = f.read()

    # HTTPレスポンスを作成
    response = {
        'statusCode': 200,
        'headers': {
            'Content-Type': 'application/zip',
            'Content-Disposition': f'attachment; filename={zip_filename}'
        },
        'body': zip_data.decode('latin1')  # バイナリデータを文字列に変換
    }

    return response

